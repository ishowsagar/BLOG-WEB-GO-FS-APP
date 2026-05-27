<!-- @ Moving forward to continuous deployment -->
IDEA - instead of deploying here and manually doing the code pushing, by using github actions, let actions access remote server, pull changes and build composed project

-> .Github/workflows/deploys.yaml - script for running cd automation

- it defines what it will do, jobs to pull latest code and deploy to vps

<!-- & Aws Ec2 -->

- This is used for spinning up an ubuntu linux distro environment where we run our cd automations, this is where we create instance of remote instances of server e.g Ubunut lts 24.0v for our project
- Connected to ubuntu server via instance created from aws, by ssh into that server public ip4 addr and granting access to it by providing the indentifier \*.pem access keys
- ssh -i {path/to/keys.pem} server@ip4 addrs
- updated packages and downloaded docker in remote server

# Remote server flow to cd

- Updated packages and installed docker.
- Update permisson for ubuntu docker user to be root user, no need of sudo permit.
- cloned project repo into /home/ubuntu dir where application repo is cloned.
- switched to project repo root
- Adding environment variables in project repo from secret and variables - adding env for instance access to this repo for operations.

# Adding secret variables of ubuntu server inot the project repo

- added SERVER_HOST - containing public ipv4 addr
- added SERVER_USER - containing "ubuntu" as instance user when we sshed the terminal <- it's the same name of the user
- added SERVER_SSH_KEY - pasted whole .pem key into it
  **Now, our project stores secrets for server address,user and access keys -> for accessing ubuntu server**

# Yaml to connect github to the server which -> log into the shh,switch dir and first push, then pull there latest code,build the project

**github actions of the repo access the server,upon pushing ssh into the terminal thorugh the already specified secrets and do the actions -by pulling latest code from the repo and deploying to the aws**

- Define name of the script pipeline
- defined what to do on pushing code to the main branch
- declared steps
  1. connect to github and access actions secrets
  2. let actions access ubuntu server by -> ssh into server using already declared secrets
  3. switch to the project repo and pull there latest code changes
  4. build in the compose by running the project repo in the aws server

1 - due to Private repo,it was failing to clone into the dir.
2 - fix the repo visibility to public from settings and successfully cloned down the repo.


<!-- ****  TESTING  **** -->
- pushing a code change -> when a change is detected -> connects to the remote ubuntu aws ec2 server instance and pull latest changes as new code was pushed -> build the project to serve in one go