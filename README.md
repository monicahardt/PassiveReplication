# PassiveReplication
# Distibuted Dictionary
To run the program you have to open up three different terminals

In the first two terminals, you have to start the three respective servers. They have to be started at port 5001 and 5002.

Feel free to copy the following three lines to each respective terminal to start the servers:

    go run server/server.go -port 5001
    go run server/server.go -port 5002

In the last terminal you have to open the frontend and the client on a port of your choosing
To do so feel free to copy the following line:

    go run client/client.go c-port 8080

It is important to wait for everything to be connected.

If you want to add a word to the dictionary write the word "add" in your client terminal, and push enter.
Then write the word you want to add to the dictionary, example: 
hello
and press enter

then write the definition to the word, example:
greeting
and press enter.
(The definition can only be one word)
Then you have added a word and definition to the dictionary.

If you want to read a defininition from the dictionary write the word "read" in your client terminal, and push enter.
then write the word you want a definition of, example:
hello
and press enter

The definition will then appear.

To crash a server use control + c on your mac keyboard
(I don't know on windows and linux)
