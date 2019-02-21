const fs = require('fs');
fs.writeFile("./test", "Hey there!", function(err) {
    if(err) {
        return console.log(err);
    }

    console.log("The file was saved!");
}); 

var stream = fs.createWriteStream("append.txt", {flags:'a'});

setTimeout(()=>{
for(var i=0; i<100000 ; i++){
    var stream = fs.createWriteStream("append.txt", {flags:'a'});
    stream.write(i + "\n");
    stream.end();
}},3000);

