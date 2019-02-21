const express = require("express");
const app = express();

http = require("http");
val = 0;

app.use((req, res, next) => {
    res.setHeader("Access-Control-Allow-Origin", "*");
    res.setHeader("Content-Type", "text/plain");
    res.setHeader(
      "Access-Control-Allow-Headers",
      "Origin, X-Requested-With, Content-Type, Accept"
    );
    res.setHeader(
      "Access-Control-Allow-Methods",
      "GET, POST, PATCH, DELETE, OPTIONS"
    );
    next();
  });

app.use((req, res, next) => {
    setInterval(function(){
        val++;
        msg = "id: msg1\ndata: test" + val + "\n\n";
        console.log(msg);
        res.write(msg);
    }, 3000);
});



module.exports = app;
