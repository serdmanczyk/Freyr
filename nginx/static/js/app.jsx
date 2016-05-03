import React from 'react';
import { render } from 'react-dom';
import { Router, Route, Link, browserHistory } from 'react-router';
import $ from 'jquery';
import moment from 'moment';

// import "css!../css/bootstrap.min.css";
// import "../css/freyr.css";
// import "../css/grayscale.css";

// var numeral = require('numeral');

// var Homepage = React.createClass({

// });

// var HeadsUp = React.createClass({
//   getInitialState: function() {
//     return {loggedin: undefined};
//   },  
// });

var LatestTable= React.createClass({
  render: function() {
    var listItems = this.props.readings.map(function(reading) {
      return (
        <tr key={reading.coreid}>
          <td>{reading.coreid}</td>
          <td>{moment(reading.posted).fromNow()}</td>
          <td>{reading.temperature.toFixed(2)}</td>
          <td>{reading.humidity.toFixed(2)}</td>
          <td>{reading.moisture.toFixed(2)}</td>
          <td>{reading.light.toFixed(2)}</td>
          <td>{reading.battery.toFixed(2)}</td>
        </tr>
      );
    }.bind(this));

    return (
      <table className="table table-hover table-bordered">
        <thead>
            <tr>
                <th>Core ID</th>
                <th>Posted</th>
                <th>Temperature</th>
                <th>Humidity</th>
                <th>Moisture</th>
                <th>Light</th>
                <th>Battery</th>
            </tr>
        </thead>
        <tbody>
          {listItems}
        </tbody>
      </table>
    )
  }
});

var Latest = React.createClass({
  getInitialState() {
      return {
        latest:[]    
      };
  },
  componentDidMount: function() {
    $.ajax({
      url: '/api/latest',
      dataType: 'json',
      cache: false,
      success: function(data) {
        data.map(function(reading) {
          reading.key = reading.coreid
        });
        this.setState({
          latest: data,
        });
      }.bind(this),
      error: function(xhr, status, err) {
        this.setState({error: err});
      }.bind(this)
    }); 
  },
  render: function() {
    if (this.state.error) {
      return (<h2> Error {this.state.error}</h2>)
    }

    return (<LatestTable readings={this.state.latest}/>)
  }
});

var Main = React.createClass({
  getInitialState: function() {
    return {loading: true};
  },
  componentDidMount: function() {
    $.ajax({
      url: '/api/user',
      dataType: 'json',
      cache: false,
      success: function(data) {
        this.setState({
          userEmail: data.email,
          userName: data.name,
          userFirstName: data.given_name,
          userLastName: data.family_name,
          loggedin: true
        });
        this.props.history.push('/latest')
      }.bind(this),
      error: function(xhr, status, err) {
        this.setState({LoggedIn: false});
      }.bind(this)
    });
  },
  render: function() {
      if (this.state.loggedin === true) {
        return (<h2> Welcome {this.state.userFirstName}</h2>)
      }
      else if (this.state.loggedin === false) {
        return (<a href="api/authorize" className="btn btn-default btn-lg">Logan</a>)
      }
      else {
        return (<img src="img/loading.svg" />)
      }
  }
});

render((
  <Router history={browserHistory} locales={['en-us']}>
    <Route path="/" component={Main} />
    <Route path="/latest" component={Latest} />
  </Router>
), document.getElementById("content"));
