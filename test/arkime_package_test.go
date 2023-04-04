package test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/shell"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestZarfPackage(t *testing.T) {
	kubeconfigPath := "/tmp/arkime_test_kubeconfig"
	gitBranch := os.Getenv("BRANCH_NAME")
	// bbPackage := os.Getenv("BIGBANG_PACKAGE_PATH")
	// testPackage := os.Getenv("TEST_PACKAGE_PATH")

	t.Log("Passed branch name: " + gitBranch)
	if gitBranch == "" {
		gitBranch = "main"
	}

	t.Log("Using branch name: " + gitBranch)
	cwd, err := os.Getwd()

	if err != nil {
		t.Error("ERROR: Unable to determine working directory, exiting." + err.Error())
	} else {
		t.Log("Working directory: " + cwd)
	}

	// Additional test environment vars. Use this to make sure proper kubeconfig is being referenced by k3d
	testEnv := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}

	clusterSetupCmd := shell.Command{
		Command: "k3d",
		Args: []string{"cluster", "create", "test-arkime",
			"--k3s-arg", "--disable=traefik@server:*",
			"--port", "0:443@loadbalancer",
			"--port", "0:80@loadbalancer",
			"--agents", "2",
			"--k3s-node-label", "arkime-capture=true@agent:0"},
		Env: testEnv,
	}

	clusterTeardownCmd := shell.Command{
		Command: "k3d",
		Args:    []string{"cluster", "delete", "test-arkime"},
		Env:     testEnv,
	}

	// if this was already running, go ahead and tear it down now.
	shell.RunCommand(t, clusterTeardownCmd)

	// to leave cluster up for examination after this run, comment this out:
	defer shell.RunCommand(t, clusterTeardownCmd)

	shell.RunCommand(t, clusterSetupCmd)

	// Identify port being used to forward to internal HTTPS
	// equivalent to: docker inspect k3d-test-arkime-serverlb --format '{{(index .NetworkSettings.Ports "443/tcp" 0).HostPort}}'
	k3dInspect := docker.Inspect(t, "k3d-test-arkime-serverlb")

	httpPort := k3dInspect.GetExposedHostPort(80)
	httpPortStr := strconv.Itoa(int(httpPort))

	httpsPort := k3dInspect.GetExposedHostPort(443)
	httpsPortStr := strconv.Itoa(int(httpsPort))

	t.Log("Using HTTP Port  " + httpPortStr)
	t.Log("Using HTTPS Port " + httpsPortStr)

	zarfInitCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"init", "--components", "git-server", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfInitCmd)

	zarfDeployDCOCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../zarf-package-dco-foundation-minimal-amd64.tar.zst", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployDCOCmd)

	// Wait for DCO elastic (Big Bang minimal deployment) to come up before deploying arkime
	// Note that k3d calls the cluster test-arkime, but actual context is called k3d-test-arkime
	opts := k8s.NewKubectlOptions("k3d-test-arkime", kubeconfigPath, "dataplane-ek")
	k8s.WaitUntilServiceAvailable(t, opts, "dataplane-ek-es-http", 40, 30*time.Second)

	zarfDeployArkimeCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../zarf-package-arkime-amd64.tar.zst", "--confirm", "--set", "BRANCH=" + gitBranch},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployArkimeCmd)

	// wait for arkime service to come up before attempting to hit it
	opts = k8s.NewKubectlOptions("k3d-test-arkime", kubeconfigPath, "arkime")
	k8s.WaitUntilServiceAvailable(t, opts, "arkime-viewer", 40, 30*time.Second)

	//-------------------------------------------------------------------------
	// Sub-tests
	//-------------------------------------------------------------------------
	// virtual service is set up as: arkime-viewer.vp.bigbang.dev
	// --fail-with-body used to fail on a 400 error which can happen when headers are incorrect.
	curlCmd := shell.Command{
		Command: "curl",
		Args: []string{"--resolve", "arkime-viewer.vp.bigbang.dev:" + httpsPortStr + ":127.0.0.1",
			"--fail-with-body",
			"-u", "localadmin:password",
			"-H", "uid: localadmin",
			"-H", "roles: arkime-user",
			"https://arkime-viewer.vp.bigbang.dev:" + httpsPortStr},
		Env: testEnv,
	}

	t.Run("Arkime runs successfully w/ initial setup", func(t *testing.T) {

		shell.RunCommand(t, curlCmd)
	})

	t.Run("Arkime undeploys cleanly", func(t *testing.T) {
		zarfDeleteArkimeCmd := shell.Command{
			Command: "zarf",
			Args:    []string{"package", "remove", "../zarf-package-arkime-amd64.tar.zst", "--confirm"},
			Env:     testEnv,
		}

		shell.RunCommand(t, zarfDeleteArkimeCmd)
	})

	t.Run("Arkime skips initial setup on re-deploy", func(t *testing.T) {
		shell.RunCommand(t, zarfDeployArkimeCmd)
	})

	t.Run("Arkime runs succesfully post initial setup", func(t *testing.T) {
		k8s.WaitUntilServiceAvailable(t, opts, "arkime-viewer", 40, 30*time.Second)
		shell.RunCommand(t, curlCmd)
	})

	//-------------------------------------------------------------------------
	// @TODO: Sensor tests
	//-------------------------------------------------------------------------
	t.Run("Arkime sensor is running", func(t *testing.T) {
		pods := k8s.ListPods(t, opts, v1.ListOptions{
			LabelSelector: "k8s-app=arkime-sensor",
		})

		for _, pod := range pods {
			t.Log("Pod log: " + k8s.GetPodLogs(t, opts, &pod, ""))
		}
	})
}
