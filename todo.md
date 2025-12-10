TODO

[] Check for working in Firefox, Chrome, Safari and Edge
[] Be complete – try to break your game and solve for these shortcomings ‘cause you know Dean will try to break them!

DONE

[x] Include a custom-token nonce on the user registration page - Generate (and store) the token on the server when the page is requested - Include information about the user (such as IP and browser), as well as a timestamp, in the token - Send the token down to the browser with the registration page - When the user submits the registration form, include the token in the request - Before creating the user on the server, validate the token: - Ensure that it matches the stored token - Ensure that user information matches what's in the request headers - Ensure that the token hasn't expired (you determine what that timeout should be) - Refer to "securitytoken.txt" for an example
[x] All queries must use prepared statements (parameterized queries)!
[x] The dice in the game board need to be more visible
[x] Remember to validate and sanitize all user input (i.e., all requests from the client)
[x] You'll use WebSockets to implement the chat (allowing for bidirectional communication) - Include at least the players' names and messages; you could also include timestamps and other info if you want
[x] Fix bug where user cannot drag token to sum of dice, and update error message to be more user friendly
[x] WebSockets and chat for individual games
[x] Missing logout button on lobby
[x] Styling for game board and auth pages should be improved
[x] It should be a bit more obvious where the board ends