# Setup

### Setting up the database

1. Install the Go pq driver `$ go get -u github.com/lib/pq`
2. Initialize the built-in SQL shell: `$ cockroach sql --insecure`
3. In the SQL shell run: 
  1. CREATE DATABASE servers_project;
  2. SET database = servers_project;
  3. CREATE TABLE "domains" (
    "name" STRING(100),
    "servers_changed" STRING(5),
    "ssl_grade" STRING(2),
    "previous_ssl_grade" STRING(2),
    "logo" STRING(150),
    "title" STRING(150),
    "is_down" STRING(5),
    "time" INT,
    PRIMARY KEY ("name")
);
4. CREATE TABLE "endpoints" (
    "name" STRING(100),
    "ip_address" STRING(100),
    "grade" STRING(2),
    "country" STRING(30),
    "owner" STRING(100),
    PRIMARY KEY ("ip_address"),
    FOREIGN KEY ("name") REFERENCES "domains"("name")
);


### Setting up the servers
1. In your terminal, go to the backend subfolder and install all Go dependencies by running `$ go get -d ./...`
2. Start the server by running  `$ go run server.go database.go domain.go`


### Setting up Vue
1. On another terminal window, go to the frontend/api subfolder and run `$ npm install`
2. Start the app by running `$ npm run serve`
3. Go to http://localhost:8081/


### Notes/Future Changes
1. Some logos will return "fakeicon.com" because the page source does not follow the standards of defining the page's logo with a rel="shortcut icon" tag.
2. The golang library fasthttp cannot handle get requests with really large headers. Thus, if you try to access a website with large headers, the package will crash
3. When trying to get the owner of an IPAddress, I am using the GoLang package "WHois." This package will sometimes return results that do not list the owner. A possible fix would be using a system call instead, finding an alternative package, or calling an external API.
4. Some websites return encrypted results to a Get request. I.E amazon.com (If using the fasthttp package)
5. If website is down or does not exist, the program returns a warning that asks the user to input another website. However, server must be restarted
6. Fasthhtp package can randomly through TCP dialing timeout errors.
