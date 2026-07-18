<!-- ! when your still is still serving old assets, you might encounter these bugs -->

1. your cache is distributed everywhere,so newer builds won't be updated untill it renews
2. You need to create invalidation -> this wipes cached resources and reloads new builds

<!-- **  solutions ** -->

> aws cloudfront create-invalidation --distribution-id E3338TWTO58MPS --paths "/\*"

<!-- ** Url mapping -->

1. Since cdn is serving application from the mainDomain cname matched to dns of the cdn, it all depands on the _route path_ pattern, if its hitting /api/\* routes -> forwarded to the backend origin and if its hitting default forwarded to the default origin dns of ec2 which is handeling frontend requests
2. for opening a ws connection url -> you need to request on the handler but stripping away the http part from the domain, so it becomes wss://domain.me/api/ws...now you may remember it is hitting /api so its sending req to backend origin server to do the job.

<!-- ! pitfals -->

1. failed to create cdn client from default config cause our default user "go-s3-client" has no policies attached to it,so explicitly had to list in the role "cdn".

<!-- Route53 mapping -->

1. Created route 53 from domain hosted zone (created from the domain itself)
2. A record is just a mapped key-val pair maps to some resource val.
3. Same way A New Record created which maps origin.domain to new ec2 instance public ip (elasticIP)
4. So,We can re-use this record to in the older distribution to map to newer instance ip.

