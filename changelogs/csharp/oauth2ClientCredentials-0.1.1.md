## oauth2ClientCredentials: 0.1.1 - 2024-09-25
### :bug: Bug Fixes
- client credentials client_id and client_secret could leak into http headers when instantiated instead of just being scoped to the token server [force-gen] *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
