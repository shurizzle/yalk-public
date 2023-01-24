# Yalk!
## Lightweight Go based chat system

### Bibliogra56op0bg  
### Requirements

#### Deployment
* Packaged and shipper in Containers (K8s or Docker)

#### Front-end <img align="center"  width="32"  height="32"  src="https://upload.wikimedia.org/wikipedia/commons/thumb/6/61/HTML5_logo_and_wordmark.svg/1200px-HTML5_logo_and_wordmark.svg.png"><img align="center"  width="32"  height="32"  src="https://cdn.iconscout.com/icon/free/png-256/css3-8-1175200.png"><img align="center"  width="32"  height="32"  src="https://i.ibb.co/02wGJvT/image.png">

##### Code
* Structure in **HTML5** 
* Styled in **CSS3** 
* Behavior in **JS** 

##### Functions

#### Back-end <img align="center"  width="64"  height="32"  src="https://i.ibb.co/SQ9hTSM/image.png">

##### Code
* Written in **Go**

##### Functions



### Bugs
- [ ] User goes offline automatically after 60 seconds
- [ ] After successful first login the cookie is deleted and user is forced back to login returning 'db sql reported connection was never closed'
- [ ] After updating profile picture, it doesn't display it again 
[//]: # Delete the current box and create a new to append from DocumentFrame ?
- [ ] Messages overflow to the right

### To-Do 

#### Importants
- [ ] Consider to move everything inside the HTTP Server folder and deem this code as the full web based app, considering Desktop and Mobile clients as separate projects

#### Importants

* Server Package
  - [ ] [Make it a systemd script](https://stackoverflow.com/questions/12486691/how-do-i-get-my-golang-web-server-to-run-in-the-background/59441983#59441983:~:text=4,service%20using%20systemd)
  - [ ] Have each user connected to instantiate a new DB connector for his database operations 
  - [ ] Make app as instance 
  - [ ] Reduce Structs and implement interfaces
  - [ ] Add statistics on current server load in HTML page
  - [ ] Separate admin settings
  - [ ] Add channel creator

* DB
  - [ ] Make all the functions to create tables

* HTML
  - [ ] Integrate profile edit features to show in main view
  - [ ] Add header with channel info
  
* CSS
  - [ ] Shrink code better, it looks digusting
  - [ ] Fix user bar row (make it proper)
  - [ ] Make decent message row
  
* Javascript
  - [ ] Create and store in broker a DB connection pointer for each user
  - [ ] Rename listen functon of broker in channel_switcher or switcher or switchboard


### To plan
- [ ] Make a manager that can instance different servers or that can manage clusters of instances in indivudal server
- [ ] Fully intergrate in a single modular package for HTTP
  - [ ] Virtual hosts
  - [ ] SSE Broker
  - [ ] Template websites
  - [ ] SQL support and connector with custom queries and generic functions
- [ ] APIs to become a separate more generic and functional package


#### Succesful deployment
