import React from 'react';
import ReactDOM from 'react-dom';
import 'bootstrap/dist/css/bootstrap.min.css';
import './index.css';
import { App } from './App';
import * as serviceWorker from './serviceWorker';

/*Since we do not have a index.html template, we create a main div where all our object will be drawn*/
const main = document.createElement('div');
document.body.appendChild(main);
main.setAttribute('id', 'main');
ReactDOM.render(<App />, main);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
