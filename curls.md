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


..

Select 
	 p.id,p.user_id,p.title,p.content,p.created_at,p.updated_at,coalesce(l.like_count,0) as like_count,u.name
	from
		posts p
	left join
    Users u
     on
    u.id = p.user_id
    left join (
        select 
            post_id,sum(like_count) as like_count
            from 
                likes l
            group by post_id
    ) l
    on l.post_id = p.id
    where 
        p.id < 9999
    order by 
        p.id desc
limit 4 ;




> comment count
- it means to fetch comment count, it need to call function passing post_id to get count
select count(*) as comments_count from comments c
    where post_id=17  
    group by post_id;

> fetch name from posts's user id 
select u.name from posts p  
    inner join users u on  
    p.user_id = u.id
    limit 7; 
    👍worked -> get all entries where ids matching, then whatever u want select from that table with dot notation