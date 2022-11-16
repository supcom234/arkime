{
    "id": "Arkime",
    "type": "chart",
    "node_affinity": "Server - Any",
    "formControls": [
        {
          "type": "textinput",
          "default_value": "assessor",
          "description": "Enter arkime user name",
          "required": true,
          "regexp": "",
          "name": "username",
          "error_message": "Enter a value"
        },
        {
          "type": "textinput",
          "description": "Enter arkime password",
          "required": true,          
          "name": "password",          
          "default_value": "Password!123456",
          "regexp": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[!@#$%^&*()<>.?])[A-Za-z\\d!@#$%^&*()<>.?]{15,}$",      
          "error_message": "Please enter a vaild password it must have a minimum of fifteen characters, at least one uppercase letter, one lowercase letter, one number and one special character.  Valid special characters !@#$%^&*()<>.?)."
        },
        {
          "type": "invisible",
          "name": "node_hostname"
        }
    ]
}
