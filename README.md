# Accessh  

**Accessh** is a simple middleman service designed to help users find the correct SSH service port based on the domain they are SSHing into. It provides a centralized hub for accessing public SSH services, ensuring users can easily navigate and connect to the right SSH based services without hassle.

## ~~Proxy22~~  

Originally my plans for this project was to make a reverse proxy for SSH connections based on the domain being connected to called *Proxy22*, but the project shifted focus after I encountered the limitations of the SSHv2 protocol (It doesn't send SNI info with client requests). Then I was planning to take user inputs and proxy the client, but attempts to implement this were met with, let’s just say, *super broken* results. So currently that implementation is on pause for now since I like this compromise I made & I'll revisit creating this the right way another day. In the meantime, Accessh evolved into something equally useful: an intuitive directory of SSH services :)

## Features  

- **Port Lookup**: Quickly identify the correct SSH port and service information for any configured service.  
- **Customizable Configuration**: Adjust settings & services via a simple JSON file.  
- **Docker-Compose Support**: Easily deploy with `docker-compose` for streamlined self-hosting.  
- **Service Directory**: Offers a list of public SSH destinations with detailed information.

## Usage  

### Running Locally

1. **Clone the Repository**:  
   ```bash  
   git clone https://github.com/your-username/accessh.git  
   cd accessh  
   ```  

2. **Modify the Configuration**:  
   Customize `config.sample.json` to suit your needs. Example:  
   ```json  
   {  
       "settings": {  
           "title": "Welcome to My SSH Lobby",  
           "description": "Enter the domain you're trying to access (type 'help' for available destinations)",  
           "SSH": {  
               "enabled": false,  
               "hostname": "0.0.0.0",  
               "port": "23"  
           }  
       },  
       "locations": {  
           "example.com": {  
               "service": "Example SSH Service",  
               "description": "For example purposes",  
               "repo": "github.com/Sample/Sample",  
               "port": 25,  
               "hostname": "example.com"  
           }  
       }  
   }  
   ```  

3. **Download dependencies**:  
   ```bash  
   go mod tidy
   ```  

4. **Run the TUI**:  
   ```bash  
   go run .
   ```  

### Self-Hosting with Docker

1. **Clone the Repository**:  
   ```bash  
   git clone https://github.com/your-username/accessh.git  
   cd accessh  
   ```  

2. **Modify the Configuration**:  
   Customize `config.sample.json` to suit your needs. Example:  
   ```json  
   {  
       "settings": {  
           "title": "Welcome to My SSH Lobby",  
           "description": "Enter the domain you're trying to access (type 'help' for available destinations)",  
           "SSH": {  
               "enabled": true,  
               "hostname": "0.0.0.0",  
               "port": "23"  
           }  
       },  
       "locations": {  
           "example.com": {  
               "service": "Example SSH Service",  
               "description": "For example purposes",  
               "repo": "github.com/Sample/Sample",  
               "port": 25,  
               "hostname": "example.com"  
           }  
       }  
   }  
   ```  

3. **Start the Service**:  
   ```bash  
   docker-compose up -d  
   ```  

## Configuration Overview  

- **`settings`**: Adjust global options like the lobby's title, description, and SSH server settings (if you plan to serve this over SSH).  
- **`locations`**: Define the available SSH services with their hostname, port, description, and repository URL.

## Future Plans  
I do plan to come back one day and make the proper reverse proxy portion of this service work but for now this will do.
