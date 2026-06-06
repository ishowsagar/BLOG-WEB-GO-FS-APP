<!-- @ Approach -->
<!--  loaded all data from getAll posts in main feed -->
<!--  set each to go to a link, which is route indexed,so whenever it goes there,it sets that /:slug being route path -->
<!-- ! declared a seperate component which get loaded when req is made to that oath, indexed one -->

<!-- docker compose env var -->

> must not have spaces in between anything when declaring the variables
> s3 client -s -> "go-s3-operator" => client for all s3 related operations just like postgres client for db
> bucket name - aws-s3-insta-bucket-storage
> docker look for env in its space as env is inaccessible to the container so when code runs it checks if they exists there.

<!-- & aws gotchas -->

1. for different sort of operation you may wanna do, like s3bucket uploads,cloudfront invalidations
2. Each requires their own sort of client types like one we create for s3 could not used to do invalidations as that is exclusively for s3bucket only.
3. When i was trying to use same s3client to do invalidations, it does not showed up callers to do that.
4. Invalidation operations are related to cloudfront only, say domain-'cdn aka cloudfront'
5. You could do all sorts of invalidations operations by accessing **Cloudfront's Client**.
6. You'd have to create client of cdfront domain and apply invalidation with that.
