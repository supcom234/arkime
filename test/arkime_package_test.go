// Test suite for arkime Zarf package. This test will use k3d to create a cluster, then use
// zarf to deploy the BB minimal package and then the arkime zarf package on top of that. It
// will connect to the viewer to ensure it started properly.
package test

import (
    "crypto/tls"
    "encoding/base64"
    "os"
    "strconv"
	"strings"
    "testing"
    "time"
    "github.com/gruntwork-io/terratest/modules/docker"
    "github.com/gruntwork-io/terratest/modules/k8s"
    "github.com/gruntwork-io/terratest/modules/shell"

    http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
)

func TestZarfPackage(t *testing.T) {
    gitBranch := os.Getenv("GITHUB_REF")
    // bbPackage := os.Getenv("BIGBANG_PACKAGE_PATH")
    // testPackage := os.Getenv("TEST_PACKAGE_PATH")

	// default creds for testing only
    auth := base64.StdEncoding.EncodeToString([]byte("localadmin:password"))

    if (gitBranch == "") {
		gitBranch = "main"
    }
    
    cwd, err := os.Getwd()

    if (err != nil){
     	t.Error("ERROR: Unable to determine working directory, exiting." + err.Error())
	} else {
        t.Log("Working directory: " + cwd)
    }

	// Additional test environment vars. Use this to make sure proper kubeconfig is being referenced by k3d
	testEnv := map[string]string{
		"KUBECONFIG": "/tmp/test_kubeconfig",
	}
    
    clusterSetupCmd := shell.Command{
  		Command: "k3d",
		Args:    []string{"cluster", "create", "test-arkime", "--k3s-arg", "--no-deploy=traefik@server:*", "--port", "0:443@loadbalancer", "--port", "0:80@loadbalancer"},
        Env:	 testEnv,
	}

    // may be helpful to comment this out when testing, to leave cluster up for examination
    clusterTeardownCmd := shell.Command{
    	Command: "k3d",
		Args:    []string{"cluster", "delete", "test-arkime"},
        Env:     testEnv,
	}

	defer shell.RunCommand(t, clusterTeardownCmd)

	shell.RunCommand(t, clusterSetupCmd)

    // Identify port being used to forward to internal HTTPS
    // equivalent to:
    //  docker inspect k3d-test-arkime-serverlb --format '{{(index .NetworkSettings.Ports "80/tcp" 0).HostPort}}'
    k3dInspect := docker.Inspect(t, "k3d-test-arkime-serverlb")
    httpPort := k3dInspect.GetExposedHostPort (80)
    httpsPort := k3dInspect.GetExposedHostPort (443)
  
    t.Log("Using HTTP Port " + strconv.Itoa(int(httpPort)))
    t.Log("Using HTTPS Port " + strconv.Itoa(int(httpsPort)))

    zarfInitCmd := shell.Command{
		Command: "zarf",
		Args:	 []string{"init", "--components", "git-server", "--confirm"},
        Env:     testEnv,
	}

	shell.RunCommand(t, zarfInitCmd)

	zarfDeployDCOCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../../zarf-package-dco-foundation-minimal-amd64.tar.zst", "--confirm"},
        Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployDCOCmd)

    // wait for DCO elastic to come up before deploying arkime
    // NewKubectlOptions (contextName, configPath, namespace) <- will need to make sure k3d either sets the context properly or manually get it n' set it
	// Note that k3d calls the cluster test-arkime, but actual context is called k3d-test-arkime
    opts := k8s.NewKubectlOptions("k3d-test-arkime", "/tmp/test_kubeconfig", "dataplane-ek");
    k8s.WaitUntilServiceAvailable(t, opts, "dataplane-ek-es-http", 10, 1*time.Minute)

	zarfDeployArkimeCmd := shell.Command{
		Command: "zarf",
		Args:	 []string{"package", "deploy", "../../zarf-package-arkime-amd64.tar.zst", "--confirm", "--set", "BRANCH=" + gitBranch, "--set", "FIRST_INSTALL=true"},
        Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployArkimeCmd)

    // wait for arkime service to come up before attempting to hit it
    opts = k8s.NewKubectlOptions("k3d-test-arkime", "/tmp/test_kubeconfig", "arkime")
    k8s.WaitUntilServiceAvailable(t, opts, "arkime-viewer", 10, 1*time.Minute)

    // Running in localhost so need to fake the hostname/cert
	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
	}

    // virtual service is set up as: arkime-viewer.vp.bigbang.dev
    httpOpts := http_helper.HttpDoOptions{
		Method:  "GET",
		Url:     "127.0.0.1:" + strconv.Itoa(int(httpsPort)),
		Body:    strings.NewReader(""),
        Headers: map[string]string{
			"Host": "arkime-viewer.vp.bigbang.dev",
			"Authorization": "Basic " + auth,
            "uid": "localadmin",
            "roles": "arkime-user",
		},
		TlsConfig: &tlsConfig,
        Timeout: 300,
    }

    // Get the arkime-viewer endpoint via istio virtual service
	_, _ = http_helper.HTTPDoWithOptions(t, httpOpts)
}

