<!-- ** Websockets ** -->

"Websockets helps us to establish two-way connection,which stays open for the exchange of information between the server and client"

`Few Things you should keep in mind before implementing ws connection from frontend side`

1. Ws connection is like connection instance opened for two-way information broadcast, it is opened on _already-defined_ url where backend handler is serving the connection from other end.
2. For connection to be successfully opened it should reach handler on the _already defined route path_;cause that would be intercepted by the handler to upgrade sent http connection into _webSocket connection_ { refer to ws handler for knowing how conn is upgraded into ws}

`Once connection is established by sending new 'ws connection' request, this is how information is shared bidirectional`

"these 4 ws-interceptors comes in action when connection is intitialized"

1. on_open -> when request sent to handler for securing a ws connection instance is successfully, this part of communication comes in action and it is ready to transmit data to the handler.
2. on_message -> this is triggered when data is sent to the client ws conn by the handler where active client's writer is writing and sending response, this part is responsible for intercepting incoming data via **event.data**.
3. on_error -> this is usually triggered when connection unable to communicate with the server;
4. on_close -> this is aggresively triggered when connection the shuts down gracefully or abruptly.

` Once both client and server are ready to share information bidirectionally, this is how sharing is done`

- Sending Data {on_open} -> data is sent to the handler via sockets.send(anyData). you could send string data,blobs,array. But for json, you must send stringified data by using JSON.stringify(anyJsonData/formData)
- Recieving Data {on_message} -> data is recieved on the clientSide ws connection in on_message block which intercepts incoming data, since when we are dealing with json form of data,we must decode incoming data with json.Parse(event.anyJsonIncomingData~data)

<!-- ! pitfalls / failures -->

<!-- ** SUCCESS ** -->

- Finally, i have implemented and learned the "gotcha" behind how ws connection is established and how data is bidirectionally exchanged
- I have tested and successfully got the backend logs that handler recieved request,checked origin and upgraded the connection.
- Once that wsConn is established on clientSide and handler upgraded the conn to store client, payload recieved successfully on the handler and able to do the rest routing logic successfully✅.
- wsConn.Send interceptor -> sends directly to the handler,when paired with ref.current to store conn, you can invoke send from anywhere to do sending✅.
