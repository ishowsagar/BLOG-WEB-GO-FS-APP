## Main.go

func newPnService function -> creates instance and starts startService func -> which runs a for loop checkign agaunbst notification chan, if there is post -> logs it to the terminal.

## noti\*\_.go

func notifyPostCreation function - runs a select group with cases for new ticker timed out rety after a time.Duration and when req's ctx done and redirects 'post' to notification chan

<!--& General flow -->

> Add a reader for that chan
> add a method which redirects output to that chan
> invoke that method on handler where it is service need to be called

<!-- @ Cache flow -->

> Add method which uses client ( which holds all the db operations) to set a key-val pair, val must be marshaled~[slice] of byte to store in cache db
> Add method to fetch that key-val pair -> result is in form of type -~ [slice] of byte -> unmanrshal into struct to return those cached vals.

## Add cache for login {try}
