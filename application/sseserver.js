SSE = require ("sse");
const app = require("./ssetest");
port = 8000;
app.set("port", port);
const server = http.createServer(app);
server.listen(port, '127.0.0.1', function(){
    var sse = new SSE(server);
    console.log("Started...");
    sse.on('connection', function(client){
        console.log("Listening");
        client.send("Hi!");
    });
});