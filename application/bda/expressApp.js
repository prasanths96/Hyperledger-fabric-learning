const express = require("express");
const bodyParser = require("body-parser");

const query = require("./query");
const enrollUser = require("./enrollUser");
const invoke = require("./invoke");
//const eventListener = require("./eventListener");

const jwt = require("jsonwebtoken");
const app = express();

app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: false }));
// Prerequisite to avoid CORS errors. (Cross-Origin-Resource-Sharing)
app.use((req, res, next) => {
    res.setHeader("Access-Control-Allow-Origin", "*");
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

loggedUser = "";
// Check authorization middleware
checkAuth = (req, res, next) => {
    try {
      // Send token in req in format: Bearer <token>
      const token = req.body.token;
      console.log("Token:"+token);
      var user = jwt.verify(token, "secret_this_should_be_longer");
      this.loggedUser = user.user;
      console.log(this.loggedUser);
      next();
    } catch (error) {
      console.log(error);
      res.status(401).json({ message: "Auth failed!" });
    }
  };

// Login/Signup API
app.post("/api/user/login", (req, res, next) => {
        // Try enroll
        const main = async function() {
            console.log("\n\n\n UserName: "+req.body.user);
            const result = await enrollUser(req.body.user, req.body.userpw);
            if(result.enrolled == true) {
                // If success, create token
                const token = await jwt.sign(
                    { user: req.body.user },
                    "secret_this_should_be_longer",
                    { expiresIn: "1h" }
                );

                res.status(200).json({
                    token: token,
                    expiresIn: 3600
                  });
            }
            else {
                res.json({
                    error: result.err
                });
            }
        }      
        main();
    });


// Application API

app.get("/api/gethistory", (req, res, next) => {
    
    const result = async function() {
        try{
            const result = await query.query();
            //console.log("Ans:" + ans);
            res.send(result);
        }
        catch(error){
            console.log("Express err:"+ error);
        }
    }
    result();
});

app.post("/api/createlayout", checkAuth, (req, res, next) => {
    const id = req.body.id;
    // const token = req.body.token;
    // const userData = jwt.verify(token, "secret_this_should_be_longer");
    // const user = userData.user;
    const address = req.body.address;
    console.log(`id: `+ id + ` address: ` + address);
    console.log('Express app: User: '+this.loggedUser);
    const main = async function(user) {
        const result = await invoke(user, 'createLayout', id, address); 
        if(result.invoked == true) {
            res.status(201).json({
                message: "Success."
            });
        }
        else {
            res.json({
                message: result.err
            });
        }       
    }
    main(this.loggedUser);
  });

  app.post("/api/requestNOC", checkAuth, (req, res, next) => {
      console.log("Express: ID: "+req.body.id);
    const id = req.body.id;
    console.log(`id: `+ id);
    const main = async function(user) {
        const result = await invoke(user, 'requestNOC', id); 
        if(result.invoked == true) {
            res.status(201).json({
                message: "Success."
            });
        }
        else {
            res.json({
                message: result.err
            });
        }       
    }
    main(this.loggedUser);
  });

  app.post("/api/viewlayout", checkAuth, (req, res, next) => {
    const id = req.body.id;
    console.log(`id: `+ id);
    const main = async function(user) {
        const result = await query(user, 'viewLayout', id); 
        if(result.success == true) {
            console.log(result.result);
            res.status(200).json({
                status: 200,
                message: "Success.",
                result: result.result
            });
        }
        else {
            res.json({
                status: 404,
                message: "Not Found.",
                result: result.err
            });
        }       
    }
    main(this.loggedUser);
  });

  app.post("/api/gethistory", checkAuth, (req, res, next) => {
    const id = req.body.id;
    console.log(`id: `+ id);
    const main = async function(user) {
        const result = await query(user, 'getHistory', id); 
        if(result.success == true) {
            console.log(result.result);
            res.status(200).json({
                status: 200,
                message: "Success.",
                result: result.result
            });
        }
        else {
            res.json({
                status: 404,
                message: "Not Found.",
                result: result.err
            });
        }       
    }
    main(this.loggedUser);
  });

  app.get("/api/events", (req, res, next) => {
    //const id = req.body.eventId;
    //console.log(`id: `+ id);
    const user = 'user1';
    const main = async function(user) {
        const result = await eventListener(user); 
        console.log("Express app EL: "+result.result);
        if(result.success == true) {
            console.log(result.result);
            res.status(200).json({
                status: 200,
                message: "Success.",
                result: result.result
            });
        }
        else {
            res.json({
                status: 404,
                message: "Not Found.",
                result: result.err
            });
        }       
    }

main(user);
});

// Test
app.use((req, res, next) => {
    res.status(200).json({
        Message: "Hello"
    });
});



// Test





module.exports = app;