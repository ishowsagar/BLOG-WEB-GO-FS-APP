<!-- @ Approach -->
<!--  loaded all data from getAll posts in main feed -->
<!--  set each to go to a link, which is route indexed,so whenever it goes there,it sets that /:slug being route path -->
<!-- ! declared a seperate component which get loaded when req is made to that oath, indexed one -->

<!-- docker compose env var -->

> must not have spaces in between anything when declaring the variables
> s3 client -s -> "go-s3-operator" => client for all s3 related operations just like postgres client for db
> bucket name - aws-s3-insta-bucket-storage
> docker look for env in its space as env is inaccessible to the container so when code runs it checks if they exists there.
