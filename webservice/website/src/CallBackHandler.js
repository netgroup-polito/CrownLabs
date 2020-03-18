import React from "react";

export default function CallBackHandler(props) {
	if(props.action === 'login')
		props.authManager.completeLogin();
	else
		props.authManager.completeLogout();
	return(<div />);
}