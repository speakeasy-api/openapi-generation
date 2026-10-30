## unions: 3.1.15 - 2026-10-28
### :bee: New Features
- accept unknown discriminator values of open discriminated unions in request input when preserveUnknownFields is enabled, so variants received from the server can be sent back; known discriminator values with an invalid payload still raise, and the Unknown fallback declines a discriminator value that another open union in the same field recognizes, so another member of the field can accept it *(commit by [@2ynn](https://github.com/2ynn))*
