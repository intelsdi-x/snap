# github.io troubleshooting tips

error:
when running `npm start`, if your node version doesn't match the one used to build will get an "Error: Missing binding ..." message

fix:
run `npm rebuild node-sass` to build bindings for current environment and then re-try `npm start`