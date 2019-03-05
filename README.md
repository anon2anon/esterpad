# Esterpad

> Fast collaborative text editor

**The project is being rewritten, it may not work, build, or whatever**
Use master branch for stable behavior

### Setting up dev environment for backend:

Install & start mongo, edit config.yaml

```bash
go get github.com/anon2anon/esterpad/cmd/...
cd $GOPATH/src/github.com/anon2anon/esterpad
$GOPATH/src/bin/esterpad config.yaml
$GOPATH/src/bin/esterpad_tester localhost:9000 2 1
```

### Setting up dev environment for frontend:

```bash
cd web

# install dependencies
npm install

# serve with hot reload at localhost:8080
npm run dev

# build for production with minification
npm run build

# lint and fix files
npm run build
```
