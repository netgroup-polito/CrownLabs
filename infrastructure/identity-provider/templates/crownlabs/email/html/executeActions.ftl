<html>
<body>
${kcSanitize(msg("executeActionsBodyHtml", user.firstName, user.lastName, user.username, user.email, link, linkExpirationFormatter(linkExpiration)))?no_esc}
</body>
</html>
