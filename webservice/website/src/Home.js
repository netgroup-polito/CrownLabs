import React, { Component } from 'react';
import {AuthenticatorInstance} from './App';
import ApiManager from './ApiManager';

let apiManager;

export default class Home extends Component {
    constructor(props) {
        super(props);
        AuthenticatorInstance.manager.getUser().then(user => {
            if(user != null) {
                apiManager = new ApiManager(user.id_token, user.token_type === undefined ? "Bearer" : user.token_type);
                //apiManager.getCRD();
                //apiManager.createCRD();
                apiManager.deleteCRD();
            }
        });
    }
    render() {
        return(<div/>);
    }
}