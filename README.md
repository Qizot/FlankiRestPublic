# FlankiRest
REST API created for a well known game called "Flanki" 

REST paths
##### User related
 - [ /user/create ](#user_create) POST
 - [ /user/login ](#user_login) POST
 - [ /user/me ](#user_me_update) PATCH
 - [ /user/me ](#user_me_delete) DELETE
 - [ /user/me ](#user_me_get) GET
##### Player related
 - [ /players ](#players) GET
 - [ /players/{id} ](#players_one) GET
 - [ /players/ranking ](#players_ranking) GET
##### Lobby's owner related
 - [ /lobbies/owner ](#lobbies_owner) GET
 - [ /lobbies/owner ](#lobbies_delete) DELETE
 - [ /lobbies/owner ](#lobbies_update) PATCH
 - [ /lobbies/owner/create ](#lobbies_create) POST
 - [ /lobbies/owner/submit ](#lobbies_submit) POST
 - [ /lobbies/owner/kick_player ](#lobbies_kick) POST
 ##### Lobby related
 - [ /lobbies ](#lobbies) GET
 - [ /lobbies/{id} ](#lobbies_get) GET
 - [ /lobbies/{id}/join ](#lobbies_join) POST
 - [ /lobbies/my ](#lobbies_my) GET
 - [ /lobbies/my/leave ](#lobbies_leave) POST
 - [ /lobbies/results ](#lobbies_results) GET
 ##### Image service endpoints
 - [ /images/{id} ](#images_get) GET
 - [ /images/my ](#images_my) GET
 - [ /images/my ](#images_my_upload) POST
 ##### Reseting password
 - [ instruction ](#reset_password)
 
 ### Web socket chat with REST endpoints
 - [ instruction ](#chat)
 
 
 

### DISCLAIMER
every function can return status 500 with json containing server internal error message
```
{
    "message": "specific server internal message"  
}
```

## API design
there are only few endpoints that doesn't require authentication via Bearer access token
- /user/create
- /user/login
- /players/ranking

all other endpoints require access token
when token has expired or is invalid http status 401 will be returned with proper json message
<br>(status code will be 500 when auth server is down)
```
{
    "message": "error message"
}
```
API doesn't have test coverage yet so bear in mind that either this README might not be up to date 
or API might be broken on it's own so ask the Creator if you find anything suspicious or you think that something can be done easier.
<br><br>

## Account 
<a name="user_create"></a>
### Creating new user
`/user/create` method POST
#### required json params
```
{
    "nickname": "should be between 4 and 20 characters",
    "email": "must be valid email format",
    "password": "must be between 6 and 32 characters",
    "sex": "either 'male' or 'female' ",
    "description": "User description from 0 up to 200 characters" // description is optional

}
```
##### response
*status code 200*
```
{
    "message": "Account has been created, you can now log in into your account"
}
```
*status code 400*
```
{
    "message": "error message"
}
```


<a name="user_login"></a>
### Logging in
`/user/login` method POST
#### required json params
```
{
    "email": "account's email",
    "password": "account's password"
}
```
#### response
*status 200*
```
{
    "access_token": "oauth generated token",
    "token_type": "Bearer",
    "expiry": "time for token to expiry in seconds, default is 7 days for a new token"
}

```
*status 400*
```
{
  "message": "error message"
}
```
*status 401*
```
{
    "message": "Invalid credentials or user doesn't exist"
}
```

<a name="user_me_update"></a>
### Modifying account information 
`/user/me` method PATCH
#### optional json params
```
{
    "nickname": "should be between 4 and 20 characters",
    "email": "must be valid email format",
    "password": "must be between 6 and 32 characters",
    "sex": "either 'male' or 'female' ",
    "description": "User description from 0 up to 200 characters" // description is optional

}
```
#### response
*status 200*
```
{
  "message": "Account has been updated"
}
```
*status 400*
```
{
  "message": "error message like invalidity of fields"
}
```

<a name="user_me_delete"></a>
### Deleting account 
`/user/me` method DELETE
<br>*no body required*
#### response
*status 200*
```
{
  "message": "Account has been deleted"
}
```
*status 400*
```
{
    "message": "account can't be deleted while user is still playing"
}
```


<a name="user_me_get"></a>
### Getting account information
`/user/me` method GET
<br>*no body required*
#### response
*status 200*
```
{
    "account": {
        "id": 2,
        "created_at": "2019-02-10T10:34:32.424339+01:00",
        "updated_at": "2019-02-10T12:03:55.36027+01:00",
        "nickname": "test1",
        "email": "test1@gmail.com",
        "sex": "male",
        "description": "test",
        "playing": false
    },
    "summary": {
        "points": 10,
        "wins": 2,
        "loses": 0
    }
}
```


## Players
<a name="players"></a>
### Getting list of all players
`/players` method GET
<br>*no body required*
#### response
*status 200*
<br> json array of players
```
[
    {
        "id": 0,
        "nickname": "",
        "sex": "",
        "description": "",
        "playing": false // is indicating if player is currently in any lobby
    },
    {
        "id": 1,
        "nickname": "",
        "sex": "",
        "description": "",
        "playing": false
    }
    ...
]
```


<a name="players_one"></a>
### Getting player by id
`/players/{id:[0-9]+}` method GET
<br>*no body required*
#### response
*status 200*
```
{
    "player": {
        "id": 2,
        "nickname": "test1",
        "sex": "male",
        "description": "test",
        "playing": false
    },
    "summary": {
        "points": 0,
        "wins": 1,
        "loses": 0
    }
}
```
*status 404*
```
{
    "message": "Player with given id has not been found"
}
```

<a name="players_ranking"></a>
### Getting players ranking
`/players/ranking` method GET
<br>*no body required*
*status 200*
```
[
    {
        "player_id": 1,
        "nickname": "test",
        "points": 0,
        "wins": 1,
        "loses": 0
    },
    {
        "player_id": 2,
        "nickname": "test1",
        "points": 0,
        "wins": 1,
        "loses": 0
    },
    ...
]
```


## Lobbies
<a name="lobbies_owner"></a>
#### Getting owner's lobby
`/lobbies/owner` method GET
<br>*no body required*
#### response
*status code 200*
```
{
    "id": 16,
    "created_at": "2019-02-06T13:05:58.7806829+01:00",
    "lobby_owner": 1,
    "name": "Best lobby ever!",
    "player_limit": 20,
    "private": false,
    "closed": false,
    "teams": [
        {
            "id": 31,
            "players": [],
            "team_color": "blue"
        },
        {
            "id": 32,
            "players": [],
            "team_color": "red"
        }
    ],
    "longitude": 48.76424343,
    "latitude": 56.9823456
}
```
*status code 404*
```
{
    "message": "Player is not an owner of any opened lobby"
}
```

<a name="lobbies_delete"></a>
### Deleting owners's lobby
`/lobbies/owner` method DELETE
<br>*no body required*
#### response
*status code 200*
```
{
    "message": "Lobby has been deleted"
}
```
*status code 400*
```
{
    "message": "error message"
}
```
*status code 404*
```
{
    "message": "error message, probably player was not an owner of any lobby"
}
```

<a name="lobbies_update"></a>
### Updating owner's lobby
`/lobbies/owner` method PATCH
#### optional json params
All limitations stay the same as for creating new lobby
```
{
    "name": "lobby name",
    "player_limit": player limit integer,
    "private": " false or true "
    "password": "required when access has changed from public to private",
    "longitude": float,
    "latitude": float
}
```
*status code 200*
```
{
    "message": "Lobby has been updated"
}
```
*status code 400*
```
{
    "message": "error message"
}
```


<a name="lobbies_create"></a>
### Creating new lobby
`/lobbies/owner/create` method POST
#### required json params
```
{
    "name": "from 4 up to 50 characters",
    "player_limit": 10, // player limit from 4 up to 20 players
    "private": "true or false",
    "password": "from 4 up to 20 characters, required only if access is private",
    "longitude": float, // will be assigned 0 if not specified
    "latitude": float
}
```
#### response
*status code 200*
<br> example of response
```
{
    "id": 16,
    "created_at": "2019-02-06T13:05:58.7806829+01:00",
    "lobby_owner": 1,
    "name": "Best lobby ever!",
    "player_limit": 20,
    "private": false,
    "closed": false,
    "teams": [
        {
            "id": 31,
            "players": [],
            "team_color": "blue"
        },
        {
            "id": 32,
            "players": [],
            "team_color": "red"
        }
    ],
    "longitude": 48.76424343,
    "latitude": 56.9823456
}
```
*status code 400*
```
{
    "message": "error message"
}
```

<a name="lobbies_submit"></a>
### Submitting match result
`/lobbies/owner/submit` method POST
#### required json params
```
{
    "winner": "either 'blue' or 'red'"
}
```
#### response
*status code 200*
```
{
    "message": "Results have been submitted"
}
```
*status code 400*
```
{
    "message": "error message"
}
```


<a name="lobbies_kick"></a>
### Kicking player out of lobby
`/lobbies/owner/kick_player` method POST
#### required json params
```
{
    "player_id": 1
}
```
*status 200*
```
{
    "message": "Player has been kicked out of lobby"
}
```
*status 404*
```
{
    "message": "error message, probably player was not found in any of lobby's teams"
}
```




<a name="lobbies"></a>
### Listing all opened lobbies
`/lobbies` method GET
<br>*no body required*
#### response
*status code 200*
```
[
    {
        "lobby_owner": 1,
        "name": "No name here",
        "player_limit": 4,
        "private": false,
        "players": 0,
        "created_at": "2019-02-10T10:32:58.347529+01:00",
        "longitude": 48.76424343,
        "latitude": 56.9823456
    },
    {
        "lobby_owner": 2,
        "name": "No name here",
        "player_limit": 4,
        "private": false,
        "players": 0,
        "created_at": "2019-02-10T10:34:52.929802+01:00",
        "longitude": 48.76424343,
        "latitude": 56.9823456
    },
    ...
]
```


<a name="lobbies_get"></a>
### Getting lobby by id
`/lobbies/{id}` method GET
<br>*no body required*
#### response
*status code 200*
<br> example of response for '/lobbies/11'
```
{
        "id": 11,
        "created_at": "2019-01-16T13:13:03.0052619+01:00",
        "lobby_owner": 3,
        "name": "best lobby ever",
        "player_limit": 6,
        "private": false,
        "closed": false,
        "teams": [
            {
                "id": 21,
                "players": [],
                "team_color": "blue"
            },
            {
                "id": 22,
                "players": [],
                "team_color": "red"
            }
        ],
        "longitude": 48.76424343,
        "latitude": 56.9823456
}
```
*status code 404*
```
{
    "message": "error message, probably lobby has not been found"
}
```

<a name="lobbies_join"></a>
### Joining lobby
`lobbies/{id}/join` method POST
#### required json params
```
{
    "team_color": "either 'blue' or 'red'",
    "password": "required when lobbie's access is set to private"
}
```
#### response
*status 200*
<br> example of response for '/lobbies/17/join'
```
{
    "message": "Joining to the lobby has been successful!"
}
```
*status 400*
```
{
    "message": "error message"
}
```
*status 401*
```
{
    "message": "Invalid lobby password"
}
```
*status 404*
```
{
    "message": "Lobby has not been found"
}
```
*status 403*
```
{
    "message": "Lobby is full"
}
```

<a name="lobbies_my"></a>
### Getting player's current lobby
`/lobbies/my` method GET
<br>*no body required*
#### response
*status code 200*
```
{
    "id": 19,
    "created_at": "2019-02-06T13:42:43.445138+01:00",
    "lobby_owner": 1,
    "name": "best lobby ever!",
    "player_limit": 20,
    "private": false,
    "closed": false,
    "teams": [
        {
            "id": 37,
            "players": [
                {
                    "player_id": 1
                }
            ],
            "team_color": "blue"
        },
        {
            "id": 38,
            "players": [],
            "team_color": "red"
        }
    ],
    "longitude": 48.76424343,
    "latitude": 56.9823456
}
```
*status code 404*
```
{
    "message": "Player was not present in any active lobby"
}
```



<a name="lobbies_leave"></a>
### Leaving lobby
`lobbies/my/leave` method POST
<br>*no body required*
#### response
*status 200*
```
{
    "message": "Left the lobby"
}
```
*status 400*
```
{
    "message": "error message"
}
```
*status 404*
```
{
    "message": "Lobby has not been found"
}
```


<a name="lobbies_results"></a>
### Listing matches results
`lobbies/results` method GET
<br>*no body required*
#### response
*status 200*
```
[
    {
        "lobby_id": 5,
        "winner": "blue",
        "teams": [
            {
                "id": 9,
                "players": [
                    {
                        "player_id": 1
                    }
                ],
                "team_color": "blue"
            },
            {
                "id": 10,
                "players": [],
                "team_color": "red"
            }
        ],
        "finished": "2019-02-05T17:36:49.821879+01:00"
    },,
    ... 
]
```

## Images service

<a name="images_get"></a>
### Getting player's avatar by  id
`images/{id}` method GET
<br>no additional header is required
#### response
*status 200*
```
binary data containing image, image type is defined in header's content type
```
*status 404*
```
{
    "message": "Image not found"
}
```

<a name="images_my"></a>
### Getting logged player's avatar
`images/my` method GET
<br>no additional header is required
#### response
*status 200*
```
binary data containing image, image type is defined in header's content type
```
*status 404*
```
{
    "message": "Image not found"
}
```

<a name="images_my_upload"></a>
### Uploading avatar
`images/my` method POST
#### Required header `Content-Type`
either `image/jpeg` or `image/png`
#### Required body
```
binary data containing image, image type must be defined in header's content type
```
#### response
*status 200*
```
{
    "message": "Image has been uploaded"
}
```
*status 400*
```
{
    "message": "error message due to invalid request"
}
```

<a name="reset_password"></a>
### Asking for password password reset email instructions
`/remember_password` method POST
#### Required body
```
{
    "email": "email onto which to sent further instructions"
}
```
#### response
*status 200*
``` 
{
    "message": "Further instructions has been sent to your email"
}
```
*status 404*
``` 
{
    "message": "account with given email has not been found"
}
```
##### Email will be sent with activation link containing reset code
 `http://domain/reset_password/{reset code}`
### Reseting password
`/reset_password` method POST
#### Required body
``` 
{
    "code": "here goes reset code",
    "new_password": "here goes new password"
}
```        
#### response
*status 200*
``` 
{
    "messsage": "Password has been changed!"
}
```
*status 400*
```
{
    "message": "either reset code is expired or password is invalid"
}
```
*status 404*
``` 
{
    "message": "account has not been found, might happen when given reset code is not assosiated with any account"
}
```




<a name="chat"></a>
## Web socket chat
Created with players' presumptive will of interacting with each other. <br>
List of endpoints available for chat interaction
- [ /chat/rooms ](#chat_rooms) GET
- [ /chat/create/{name} ](#char_create) POST
- [ /chat/close/{name} ](#chat_close) POST
- [ /chat/join/{name} ](#chat_join) GET

##### Authorization
Users will be by default authorized with Flanki API, that means that OAUTH tokens will be required(except for /chat/rooms and /chat/join). <br><br>

##### Default chat room
Only one consistent chat is ensured to be available at any time with a name `general` availble at the endpoint `/chat/join/general`. <br><br>

##### Making web socket connection
Web socket connection can only be established via `/chat/join/{name}` endpoint. Request will be upgraded to web socket provided that the join request was sent to `ws://domain/chat/join/{name}` <br><br>

<a name="interact"></a>
##### To start interacting
Before sending any messages user is required to send authorization message
containing his auth token. If the token happens to be invalid, connection will be immediately closed.
After sending token, user is allowed from now to send and receive JSON messages of format specified below.  
```
{
    "token": "token received from Flanki API"
}
```

#### Web socket message structure
###### *DISCLAIMER*
if you happen to send unknown actions to the server 
or will you be unauthorized to do certain types of actions,
nothing will happen, no error message will be sent back to you. 
```
// Message that can be sent
{
    "action": "one of ['message', 'users', 'kick']",
    "text": "either text message for 'message' action or user's nickname to be kicked with 'kick'
}
```
```
// Message that can be received
{
    "nickname": "nickname of message creator",
    "action": "either 'message' with text message or 'kick' meaning that you've been kicked out (after this message socket connection will be closed)",
    "text": "text message",
    "time": "server side time when message has been processed",
    "data": "might be null if message didn't contain any additional data, but can be an array of nicknames for `users` action e.g. ['user1', 'user2',...]"
}
```
<a name="chat_rooms"></a>
 ### Listing available chat rooms
 `/chat/rooms` GET
 <br>*no body required*
 #### response
 *status 200*
 ```
 ["room1", "room2",...]
```

<a name="chat_create"></a>
### Creating new chat room
`/chat/create/{name}` POST
##### Required `Authorization` header with Flanki API Oauth token
After request user for now is stated as the room's owner.
He or she will be able to kick players out of the room an will become
the only person to be able to close the room (excluding admins).
<br>
Kicking players out can be done with a JSON message sent via web socket.
<br>
Endpoint creates new room with given `name`
#### response
*status 200*
```
{
    "message": "created new chatroom"
}
```
*status 400*
```
{
    "message": "rooms limit reached or lobby's name is already taken"
}
```

<a name="chat_close"></a>
### Closing chat room
`/chat/close/{name}`
##### Required `Authorization` header with Flanki API Oauth token
User has to be the owner of the chat room to be able to close it
#### response
*status 200*
```
{
    "message": "room has been closed"
}
```
*status 401*
```
{
    "message": "user was not an owner of given room"
}
```
*status 404*
```
{
    "message": "room not found"
}
```

<a name="chat_join"></a>
### Joining chat room
`/chat/join/{name}` GET
<br>
this is the only endpoint that will upgrade to web socket.
<br><br>
`if the connection was successful, web socket should be established.`
<br>
`Connection will be immediately closed if room with given name was not found`
<br><br>
Everything else connected with chat interaction has been explained in [ interaction section ](#interact).








