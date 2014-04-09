package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	portFlag       = flag.Int("port", 8100, "Port on which to listen for HTTP requests.")
	addrFlag       = flag.String("address", "", "Address on which to listen for HTTP requests.")
	mqURIFlag      = flag.String("mqURI", "", "[REQUIRED] URI of the Rabbit MQ. (e.g. amqp://<host>:<port>)")
	exchangeFlag   = flag.String("exchange", "", "Rabbit MQ exchange.")
	routingKeyFlag = flag.String("routingKey", "somequeue", "Rabbit MQ routingKey.")
    exchangeType   = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
)

func usage() {
	flag.Usage()
	os.Exit(1)
}

func processRequest(req *http.Request) (err error) {
	buffer, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}
	connection, err := amqp.Dial(*mqURIFlag)
	if err != nil {
		return
	}
	channel, err := connection.Channel()
	if err != nil {
		return
	}
    err = channel.ExchangeDeclare(
        *exchangeFlag, // name
        *exchangeType, // type
        true,         // durable
        false,        // auto-deleted
        false,        // internal
        false,        // noWait
        nil,          // arguments
    )
	if err != nil {
		return
	}
	channel.Publish(
		*exchangeFlag,
		*routingKeyFlag,
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     req.Header.Get("Content-Type"),
			ContentEncoding: "UTF-8", //TODO change
			Body:            buffer,
			DeliveryMode:    amqp.Persistent,
			Priority:        0,
		},
	)
	err = connection.Close()
	return
}

func MyServer(w http.ResponseWriter, req *http.Request) {
	var err error
    respCode := http.StatusNotImplemented
    respStatus := "sorry"
    log.Printf("method %s", req.Method)
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		err = processRequest(req)
        if err != nil {
            respCode = http.StatusInternalServerError
            respStatus = err.Error()
        } else {
            respCode = 200
            respStatus = "OK"
        }
    }
    if respCode == 200 {
        io.WriteString(w, "OK\n")
    } else {
        http.Error(w, respStatus, respCode)
    }
	log.Printf("%s %s %s %s %s %d %s", req.RemoteAddr, req.Header.Get("User-Agent"), req.Method, req.RequestURI, req.Proto, respCode, respStatus)
}

func main() {
	flag.Parse()
	if *mqURIFlag == "" {
		usage()
	}
    http.HandleFunc("/", MyServer)
    log.Printf("Starting server on %s:%d.", *addrFlag, *portFlag)
    err := http.ListenAndServe(fmt.Sprintf("%s:%d", *addrFlag, *portFlag), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
