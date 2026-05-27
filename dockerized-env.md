<!-- @ Moving forward to continuous deployment -->

-> .Github/workflows/deploys.yaml - script for running cd automation
- it defines what it will do, jobs to pull latest code and deploy to vps


<!-- & Aws Ec2 -->
- This is used for spinning up an ubuntu linux distro environment where we run our cd automations, this is where we create instance of remote instances of server e.g Ubunut lts 24.0v for our project
- Connected to ubuntu server via instance created from aws, by ssh into that server public ip4 addr and granting access to it by providing the indentifier *.pem access keys 
- ssh -i {path/to/keys.pem} server@ip4 addrs
- updated packages and downloaded docker in remote server
