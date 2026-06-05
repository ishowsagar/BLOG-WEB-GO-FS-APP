<!-- ! when your still is still serving old assets, you might encounter these bugs -->

1. your cache is distributed everywhere,so newer builds won't be updated untill it renews
2. You need to create invalidation -> this wipes cached resources and reloads new builds


<!-- **  solutions ** -->
> aws cloudfront create-invalidation --distribution-id E3338TWTO58MPS --paths "/*"