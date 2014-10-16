<div>Vocabulary is a resource to learn foreign language words.</div>

<div>X-Powered by Go on the server side, Mongo DB to store data and Knockout JS on the front end.</div> 

<div>To run the application you need to install Go (golang) from https://golang.org/doc/install</div> 
<div>Setup $GOROOT and $GOPATH to be able to download third party packages.</div>

Download and run Mongo DB server.

To use JSON Web Token generate public and private keys using the command:<br />
  <indent>$ openssl genrsa -out demo.rsa 1024 # the 1024 is the size of the key we are generating</indent>
  $ openssl rsa -in demo.rsa -pubout > demo.rsa.pub 
