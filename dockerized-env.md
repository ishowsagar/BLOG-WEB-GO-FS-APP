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
- added `restart: unless-stopped` to let aws acknowledge that whenever instance is booted up -> boot up these, unless explicitly not told to stop -> serves application as soon as vps is up⚡

<!-- ! failure -->

1. Should never explicitly define env variables like - xName=yVal -> this would break terms
   fix - instead let it pull from local machine env by telling the placeholder it stores in the actual .env file e.g -> xName=${xName} -> it fixes and tells to pull env locally instead of hardcoded stored
2. build crash and backend container exited out of the composed environment cause there was env failure, as i had mapped env vars in the compose but also gave permission to use local machine env, so both places env makes it hang up unexpectedly
3. Trailing spaces are not allowed in the env,cleared those unwanted spaces in the env delcarement.
4. frontend failed to load on the port 5173 cause instance did not expose port [ :5173 ] in inbound rules, once set and save rule -> ready to serve that
5. since we moved entire project to dockerized env and running there, we have to map base origin of the instance to the api calls
6. Also again, classic clean state postgresDB running, had to backup original data and dump into the docker postgres container
7. Keys get compromised if mistakenly pushed to the compose env, makes it inactive
8. If keys get exposed and restricted, make sure to create new , grab both accessid and secretkey, replace them in env only, don't expose to compose env of any service
9. Upon restarting ec2 instance, Don't forget to change newly assgined ipv4 addr cause it changes everytime, make sure to update in the github secrets.

<!-- ** Success ** -->

1- Env fixes -> maps out correct variables to codebase placeholders which was being picked by native codefile
2- CORS & URL fixes -> maps out correct url to api calls base url, from localhost to public ipv4 addr of ec2
