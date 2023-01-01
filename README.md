# PassiveReplication

To run the program you have to open up four different terminals

In the first three terminals, you have to start the three respective servers. They have to be started at port 5001, 5002 and 5003.

Feel free to copy the following three lines to each respective terminal to start the servers:

    go run server/server.go -port 5001
    go run server/server.go -port 5002
    go run server/server.go -port 5003

In the last terminal you have to open the frontend and the client on a port of your choosing
To do so feel free to copy the following line:

    go run client/frontend.go client/client.go c-port 8080

It is important to wait for everything to be connected.
Write in the exam that we could lock things to make sure that you cannot continue until everything is done

We have chosen that it is the server with the highest portnumber that is always chosen to be the leader.
This also makes it easier in the frontend, where is there is a leadercrash we just delete the server with the higest portnumber
knowing that the leader MUST have been that one.
