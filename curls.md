tokens testing -
{
"status": "login successfull",
"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcnkiOiIyMDI2LTA0LTMwVDEyOjQwOjUzLjgxMjExNTkrMDU6MzAiLCJ1c2VySUQiOjJ9.owtZPDUeH8aY-ROX3kZqxhJN-HT4dtnmQMVCoF8YpC4",
"user": {
"id": 2,
"name": "protected_User",
"email": "protecteduser@gmail.com",
"created_at": "2026-04-29T11:39:25.843677+05:30"
}
}

user => {
"name" : "denverJR",
"email" : "denverjr@gmail.com",
"password": "denverJR"
}

token => {
"status": "login successfull",
"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcnkiOiIyMDI2LTA0LTMwVDIwOjEyOjI0LjM3NjYzMzYrMDU6MzAiLCJ1c2VySUQiOjN9.tMs5BuzIjxEa40LXLicCsKOadqZF_zRLUkY1ZuKjYdY",
"user": "denverJR"
}


# check expiry for this user
**client** 
> {
"name" : "ayush",
"email" : "ayush@gmail.com",
"password": "ayush"
}

**token** time- 04:28 pm
> {
    "status": "login successfull",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcnkiOiIyMDI2LTA1LTAxVDExOjM3OjEwLjgzNTk2NTQrMDU6MzAiLCJ1c2VyX2lkIjo0fQ.kF06bIsDIw7aOpwa0fZrIjn2GTUV8CgBVRtdPQ0nY30",
    "user": "ayush"
}

**Post**
> {
    "Code": 200,
    "Status": "post created successfully ✅",
    "Post": {
        "id": 1,
        "user_id": 4,
        "title": "Intro - first port of the blog ever",
        "content": "i've created my first post, feeling exicted!.",
        "created_at": "2026-04-30T18:49:37.681846+05:30",
        "updated_at": "2026-04-30T18:49:37.681846+05:30"
    }
}