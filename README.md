# cardamomo


### Installation

Import the next package in your project:

```sh
import (
    "github.com/zombcode/cardamomo"
)
```

while executing this in a terminal:

```sh
$ go get "github.com/zombcode/cardamomo"
```

Or if you prefer, you can clone the repository with:

```sh
$ git clone https://github.com/zombcode/cardamomo.git
```

### First steps

#### HTTP

##### Basics

In order to instanciate the **cardamomo** framework in your project, you must write:

```sh
c := cardamomo.Instance("8000") // 8000 is the port
```

At this moment your cardamomo are instanciated. When you are ready, you can do:

```sh
c.Run()
```

to run the cardamomo http server.

##### GET

In order to generate GET patterns you can do:

```sh
c.Get("/", func(req cardamomo.Request, res cardamomo.Response) {
    res.Send("Hello world!");
})
```

As you can see, the variable **res** in the callback function is used to send data to the client.

##### POST

In order to generate POST patterns you can do:

```sh
c.Post("/post", func(req cardamomo.Request, res cardamomo.Response) {
    res.Send("Hello /post!");
})
```

As you can see, the variable **res** in callback function is used to send data to the client.

##### BASE

The **base** is used to GROUP various routes below the same base path.

```sh
c.Base("/base", func(router *cardamomo.Router) {
    router.Get("/route", func(req cardamomo.Request, res cardamomo.Response) {
        res.Send("Hello route base/route!");
    })
})
```

##### PARAMETERS

If you want to send parameters you can do it with:

```sh
c.Post("/post", func(req cardamomo.Request, res cardamomo.Response) {
    foo := req.GetParam("foo", "default value") // This should be return your param
    res.Send("Hello /post!");
})
```

This example use the **POST** method but you can use **GET** is the same way.

Otherwise, you can send parameters inside the **URL** with:

```sh
c.Get("/routeget2/:param1/and/:param2", func(req cardamomo.Request, res cardamomo.Response) {
  res.Send("Hello route get 1 with param1 = " + req.GetParam("param1", "default value") + " and param2 = " + req.GetParam("param2", "default value") + "!");
})
```

In this example you can use the **:param1** and **:param2** as variables. In order to test this use:

```sh
http://localhost:8000/routeget2/theparameter1/and/theparameter2
```

In the response you can see:

```sh
Hello route get 2 with param1 = theparameter1 and param2 = theparameter2!
```

If you use JSON for sending parameters to the server, you can use

```sh
req.GetParamJSON("username", cardamomo.JSONC{}).(string)
```

for example for obtain a parameter "username" that is a "string" with a default
empty JSON.

##### REGEX

You can use REGEX into routes, for that you need to insert the regex between "{{theregex}}", for example:

```sh
c.Get("/routeget3/{{a([a-zA-Z0-9]+)b$}}", func(req cardamomo.Request, res cardamomo.Response) {
  res.Send("Hello! This route uses REGEX! Only URL that use parameters between 'a' and 'b'");
})
```

With this lines, you can catch all request with the pattern "/routeget3/" followed with characters
between 0-9, a-z, A-Z enclosed with "a" and "b" characters.

##### JSON Responses

If you need **JSON** formatted responses for your **REST API**,
you can do it like this:

```sh
type Box struct {
  Size BoxSize
  Color  string
  Open   bool
}

type BoxSize struct {
  Width  int
  Height int
}

...

c.Get("/routejson", func(req cardamomo.Request, res cardamomo.Response) {
  boxsize := BoxSize {
    Width:  10,
    Height: 20,
  }

  box := Box {
    Size: boxsize,
    Color:  "blue",
    Open:   false,
  }

  res.SendJSON(box)
})
```

or if you prefer, you can use the object "JSONC" stored in the cardamomo package.

```sh
cardamomo.JSONC{
  "foo": "bar",
  "bar": "foo",
}
```

##### Cookies

If you need cookies, you can do this in order to add a new cookie:

```sh
c.Get("/setcookie/:key/:value", func(req cardamomo.Request, res cardamomo.Response) {
  key := req.GetParam("key", "")
  value := req.GetParam("value", "")

  expire := time.Now().AddDate(0, 0, 1) // Expires in one day!
  req.SetCookie(key, value, "/", "localhost", expire, 86400, false, false) // key, value, path, domain, expiration, max-age, httponly, secure

  res.Send("Added cookie \"" + key + "\"=\"" + value + "\"");
})
```

and if you need to get a cookie:

```sh
c.Get("/getcookie/:key", func(req cardamomo.Request, res cardamomo.Response) {
  key := req.GetParam("key", "")

  cookie := req.GetCookie(key, "empty cookie!"); // key, defaultValue

  res.Send("The value for cookie \"" + key + "\" is \"" + cookie + "\"");
})
```

Finally, if you need to delete a cookie you can do this:

```sh
c.Get("/deletecookie/:key", func(req cardamomo.Request, res cardamomo.Response) {
  key := req.GetParam("key", "")

  req.DeleteCookie(key, "/", "localhost"); // key, path, domain

  res.Send("Deleted cookie \"" + key + "\"");
})
```

#### Sockets

In order to start a socket, you need to instanciate an **HTTP** server in the
first place.

```sh
c := cardamomo.Instance("8000")
```

After that you can now create the **WebSocket** server with:

```sh
socket := c.OpenSocket()
```

and use the **socket** variable or whatever you want to control your new socket.

##### Socket base

The base works the same way as the **HTTP** base:

```sh
socket.OnSocketBase("/base1", func(client *cardamomo.SocketClient) {
  // Write your code here!
})
```

This event is called whenever a new client is connected using the path "base1":

##### Socket base actions

```sh
socket.OnSocketBase("/base1", func(client *cardamomo.SocketClient) {
  client.OnSocketAction("action1", func(sparams map[string]interface{}) {
    // Write your code here!
  })
})
```

The actions is called when the client sends a message using the required JSON format:

> - Action: A string containing the action name
> - Params: A JSON containing the params that will be sent to the action

##### Socket send

This action is used with the client variable to send data to the client websocket.
For example, you can send a message like this:

```sh
type MessageParam struct {
  param1 string
  param2 string
}

socket.OnSocketBase("/base1", func(client *cardamomo.SocketClient) {
  client.OnSocketAction("action1", func(sparams map[string]interface{}) {
    params := MessageParam{param1: "The param 1", param2: "The param 2"}
    client.Send("action1", params)
  })
})
```

If you need to send a **broadcast** to all the clients, clients attached to a concrete base
or a concrete client, you can do this:

To broadcast:
```sh
...

socket.Send("theaction", theparams)

...
```

To a concrete base:
```sh
...

socket.SendBase("/thebase", "theaction", theparams)

...
```

To a concrete client:
```sh
...

socket.SendClient("theclientID","theaction", theparams)

...
```

##### Socket cluster

If you need scalability in your system and you use sockets, THIS IS YOUR DAY! You can use
the cluster configuration to share sockets between servers. In order to do that, you can do this:

```sh
socket := c.OpenSocket()
socket.Cluster(cardamomo.SocketClusterParams{ // You can use this lines for cluster testing
  Hosts: []cardamomo.SocketClusterHost{ // Write a list of ALL servers included in the cluster
    cardamomo.SocketClusterHost{
      Host: "192.168.0.214", // Use the server 1 IP
      Port: "8000", // Use the server 1 PORT
      Master: true, // Only ONE server can be MASTER
    },
    cardamomo.SocketClusterHost{
      Host: "192.168.0.214", // Use the server 2 IP
      Port: "8001", // Use the server 2 PORT
      Master: false, // Only ONE server can be MASTER
    },
  },
  Password: "examplepass",
})
socket.SendClient("as12df34gh56", "testing", cardamomo.JSONC{"foo":"bar"}); // This is for testing the communication between sockets, you can use a real client ID for that
```

In this example you can connect to SERVER 1 and use SendClient with SERVER 2 user id to send
info between servers.

Go to **cardamomo-examples** for more info about sockets.

#### Error handler

You can debug errors with the error handler using:

```sh
c.SetErrorHandler(func (code string, req cardamomo.Request, res cardamomo.Response) {
  fmt.Printf("\nError: %s\n", code)

  if( code == "404" ) {
    res.Send("Error 404!");
  }
})
```

For example, you can use this to send 404 errors.

##### In future

At this moment the framework is very simple. In the future we want to implement:

> - Layout manager
> - File upload (single and multiple)

### Version
**Beta - 1.0.0**
