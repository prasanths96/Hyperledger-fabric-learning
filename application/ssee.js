var express = require('express');
var app = express();
var eventPayload = require("./bda/eventListener");
var route = express.Router();
 
var connections = [];
route.get('/data', function (req, res) {
req.socket.setTimeout(50000000);
res.writeHead(200, {
'Content-Type':'text/event-stream',
'Cache-Control':'no-cache',
'Connection':'keep-alive',
"Access-Control-Allow-Origin": "*"
})
res.write('/n');
connections.push(res)
console.log(res);
 
 
req.on("close", function () {
var rem=0;
for (let i=0; i<connections.length; i++) {
if (connections[i] ==res) {
rem=i;
break;
}
}
connections.splice(rem, 1)
 
});
 
 
})
 val = 0;
setInterval(function () {
    val ++;
connections.forEach(function (res) {
var d=new Date();
res.write(`data: ${eventPayload}\n\n`);
console.log(val);
})
}, 1000);
app.use('/api', route)
app.listen(8000, function () {
console.log("Listening on port 8000")
})