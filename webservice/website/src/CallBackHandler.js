import React, { Component } from 'react';
import {AuthenticatorInstance} from './App';

export default class CallBackHandler extends Component {
    constructor(props) {
		super(props);
		AuthenticatorInstance.completeLogin();
	}
    
    render() {
	    return(<div />);
	}
}