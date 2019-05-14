![](https://cl.ly/1m341B1l0l2P/plank_logo-final.png)

# plank

Spinnaker SDK for Go.

*Plank is a work in progress and will change drastically over the next few months.*

## What is it?
A package used by services that interact with Spinnaker's micro-services. It is not intended to be a client which interacts with Spinnaker's outward facing API.

## Why is it named plank?
Because it's funny.

## How do I use it?
Very carefully. :smiley:

Basic concept is that you instantiate a Plank client thusly:

```go
client := plank.New()
```

If you'd like to supply your own http client (we use a pooled client by default):
```go
client := plank.New(plank.WithClient(&http.Client{}))
```

To tune the maxiumum number of retries or set an exponential backoff value:
```go
client := plank.New(plank.WithMaxRetries(5), plank.WithRetryIncrement(5 * time.Second))
```

You can (or may need to) replace the base URLs for the microservices by
assign them to the keys in the `client.URLs` map:

```go
client.URLs["orca"] = "http://my-orca:8083"
client.URLs["front50"] = config.Front50.BaseURL
```

After that, you just use the Plank functions to "do stuff":

```go
app, err := client.GetApplication("myappname")
pipelines, err := client.GetPipelines(app.Name)
// etc...
```


