## REST-API WITH AUTHENTICATION

# Build with GO & GIN and Mongodb

# Features
1. User signup
2. User signin
3. Only authenticated user can access /users/:userid route
4. only authenticated admin can access /users route


[GIN] 2022/07/14 - 13:43:16 | 200 |    2.1871188s |       127.0.0.1 | POST     "/users/signup"
[GIN] 2022/07/14 - 13:43:47 | 200 |    2.1329916s |       127.0.0.1 | POST     "/users/signup"
[GIN] 2022/07/14 - 13:44:04 | 200 |     2.191119s |       127.0.0.1 | POST     "/users/login"
[GIN] 2022/07/14 - 13:44:21 | 200 |       50.77ms |       127.0.0.1 | GET      "/users"
[GIN] 2022/07/14 - 13:45:06 | 200 |      1.6149ms |       127.0.0.1 | GET      "/users/62ce323bbde3ae7a5e3d63b8"