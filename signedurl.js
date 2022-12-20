const express = require('express')
const app = express()
const fs = require('fs');

const config = {
   secret: "My secret was here onetime", // used to sign and verify everything *required*
   key: 'hash' // the query string key to use (defaults to 'hash')
};

const options = {
   method: 'GET', // request method (defaults to 'GET')
   ttl: 3600 // expiry time in seconds (optional)
};

const signer = require('signed-url')(config);
const url = 'http://194.61.28.250:3535/download?file=data&ext=pdf'
const url2 = 'http://194.61.28.250:3535/download?file=nonexistant&ext=pdf'

app.get('/list', (req, res) => {

 const signedUrl = signer.sign(url, options);
 const signedUrl2 = signer.sign(url2, options);
 res.send({"url":signedUrl,"url2":signedUrl2})
})

app.get('/download', (req, res) => {
 
 const fullLink = "http://194.61.28.250:3535" + req.originalUrl;
 const complete_filename = req.query.file + "." + req.query.ext;
 const valid = signer.verify(fullLink, options);
 if(!valid){
  res.send('Tampering Detected!')
  return
 }

 res.download(complete_filename)
})

app.listen('3535')
