<html>
<body>
${kcSanitize(msg("passwordResetBodyHtml", user.firstName, user.lastName, link, linkExpirationFormatter(linkExpiration)))?no_esc}
</body>
</html>