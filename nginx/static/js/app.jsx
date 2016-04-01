import React from 'react';
import { render } from 'react-dom';
import { Router, Route, Link, browserHistory } from 'react-router';
import $ from 'jquery';
// import moment from 'moment';

// import "css!../css/bootstrap.min.css";
// import "../css/gardenspark.css";
// import "../css/grayscale.css";

// var numeral = require('numeral');

// var Homepage = React.createClass({

// });

var Main = React.createClass({
  getInitialState: function() {
    return {loggedin: undefined};
  },
  componentDidMount: function() {
    $.ajax({
      url: '/api/',
      dataType: 'text',
      cache: false,
      success: function(data) {
        this.setState({loggedin: true});
      }.bind(this),
      error: function(xhr, status, err) {
        this.setState({loggedin: false});
      }.bind(this)
    });
  },
  render: function() {
    return (function(state) {
      if (this.state.loggedin === true) {
        return (<h2> Welcome </h2>)
      }
      else if (this.state.loggedin === false) {
        return (<a href="api/authorize" className="btn btn-default btn-lg">Login</a>)
      }
      else {
        return (<img src="img/loading.svg" />)
      }
    }.bind(this))(this.state.loggedin);
  }
});

render((
  <Router history={browserHistory} locales={['en-us']}>
    <Route path="/" component={Main} />
  </Router>
), document.getElementById("content"));
