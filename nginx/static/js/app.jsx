import React from 'react';
import { render } from 'react-dom';
import { Router, Route, Link, browserHistory } from 'react-router';
import $ from 'jquery';
import d3 from 'd3';
import ReactFauxDOM from 'react-faux-dom';
import DateRangePicker from 'react-bootstrap-daterangepicker';
import moment from 'moment';
import _ from 'underscore'
import { Button, Glyphicon } from 'react-bootstrap';

var ReadingsTable = React.createClass({
  clickedCoreRow(key) {
    return (e) => {
      e.preventDefault();
      if (this.props.nav) {
        browserHistory.push('/readings/' + key);
      }
    };
  },
  render() {
    var listItems = this.props.readings.map(reading => {
      return (
            <tr key={reading.coreid + reading.posted} onClick={this.clickedCoreRow(reading.coreid)}>
              <td>{reading.coreid}</td>
              <td>{moment(reading.posted).fromNow()}</td>
              <td>{reading.temperature.toFixed(2)}</td>
              <td>{reading.humidity.toFixed(2)}</td>
              <td>{reading.moisture.toFixed(2)}</td>
              <td>{reading.light.toFixed(2)}</td>
              <td>{reading.battery.toFixed(2)}</td>
            </tr>
      );
    });

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

var ReadingsChart = React.createClass({
  getInitialState() {
    var sod = moment().startOf('day');
    var eod = moment().endOf('day');
    var som = moment().startOf('month');
    var eom = moment().endOf('month');
      return {
        ranges: {
          'Today': [sod, eod],
          'Yesterday': [sod.clone().subtract(1, 'days'), eod.clone().subtract(1, 'days')],
          'Last 7 Days': [sod.clone().subtract(6, 'days'), eod],
          'Last 30 Days': [sod.clone().subtract(29, 'days'), eod],
          'This Month': [som, eom],
          'Last Month': [som.clone().subtract(1, 'month'), eom.clone().subtract(1, 'month')]
        },
        startDate: sod,
        endDate: eod,
        readings:[],
        readingType: "temperature",
        windowInnerWidth: window.innerWidth,
        windowInnerHeight: window.innerHeight,
      };
  },
  componentDidMount() {
    window.addEventListener("resize", this.updateDimensions);
    this.fetchReadings(this.state.startDate, this.state.endDate)
  },
  updateDimensions() {
    this.setState({
      windowInnerWidth: window.innerWidth,
      windowInnerHeight: window.innerHeight,
    });
  },
  handleTypeChange(e) {
    this.setState({
      readingType: e.target.value,
    });
  },
  componentWillUnmount() {
    window.removeEventListener("resize", this.updateDimensions);
  },
  fetchReadings(startDate, endDate) {
    $.get('/api/readings', {
        'start': startDate.format(),
        'end': endDate.format(),
        'core': this.props.params.coreid
      },
      'json')
        .done(data => {
          this.setState({
            startDate: startDate,
            endDate: endDate,
            readings: _.sortBy(data, 'posted')
          });
        })
        .fail(err => {
          this.setState({
            readings: [],
          });
          console.log(err);
          // this.setState({error: err});
        });
  },
  onApply(event, picker) {
    this.fetchReadings(picker.startDate, picker.endDate);
  },
  render() {
    var data = this.state.readings;
    var type = this.state.readingType;
    var margin = {top: 20, right: 20, bottom: 30, left: 50};
    var width = this.state.windowInnerWidth * 0.60;
    var height = this.state.windowInnerHeight * 0.57;
    var minWidth = 100;
    var minHeight = 100;

    width = width > minWidth ? width : minWidth;
    height = height > minHeight ? height : minHeight;

    width = width - margin.left - margin.right
    height = height - margin.top - margin.bottom

    var x = d3.time.scale()
    .range([0, width])

    var y = d3.scale.linear()
    .range([height, 0])

    var xAxis = d3.svg.axis()
    .scale(x)
    .orient('bottom')
    .ticks(7)

    var yAxis = d3.svg.axis()
    .scale(y)
    .orient('left')

    data.forEach(function (d) {
      d.timestamp = moment(d.posted)
    })

    var line = d3.svg.line()
    .x(function (d) { return x(d.timestamp) })
    .y(function (d) { return y(d[type]) })


    var node = ReactFauxDOM.createElement('div')
    var svg = d3.select(node)   
       .classed("svg-container", true) //container class to make it responsive
       .append("svg")
       .classed("freyr-chart", true)
        .attr('width', width + margin.left + margin.right)
        .attr('height', height + margin.top + margin.bottom)
        .append('g')
        .attr('transform', 'translate(' + margin.left + ',' + margin.top + ')')

    x.domain(d3.extent(data, function (d) { return d.timestamp }))
    y.domain(d3.extent(data, function (d) { return d[type] }))

    svg.append('g')
      .attr('class', 'x axis')
      .attr('transform', 'translate(0,' + height + ')')
      .call(xAxis)

    svg.append('g')
      .attr('class', 'y axis')
      .call(yAxis)
      .append('text')
      .attr('transform', 'rotate(-90)')
      .attr('y', 6)
      .attr('dy', '.5em')
      .style('text-anchor', 'end')
      .text(this.state.readingType)

    svg.append('path')
      .datum(data)
      .attr('class', 'line')
      .attr('d', line)
    
    var start = this.state.startDate.format('YYYY-MM-DD');
    var end = this.state.endDate.format('YYYY-MM-DD');
    var label = start + ' - ' + end;
    if (start === end) {
      label = start;
    }

    return (
      <div>
          <DateRangePicker startDate={this.state.startDate} endDate={this.state.endDate} ranges={this.state.ranges} onApply={this.onApply}>
            <Button className="selected-date-range-btn">
              <div className="pull-left"><Glyphicon glyph="calendar" /></div>
              <div className="pull-right">
                <span>
                  {label}
                </span>
                <span className="caret"></span>
              </div>
            </Button>
          </DateRangePicker>
          <select className="btn btn-default btn-sm" onChange={this.handleTypeChange}>
            <option value="temperature">Temperature</option>
            <option value="humidity">Humidity</option>
            <option value="moisture">Moisture</option>
            <option value="light">Light</option>
            <option value="battery">Battery</option>
          </select>
          {node.toReact()}
      </div>
    );
    // return <span />
  },
});

var Latest = React.createClass({
  getInitialState() {
      return {
        latest:[]
      };
  },
  componentDidMount() {
    $.get('/api/latest', null, 'json')
      .done(data => {
        this.setState({
          latest: data,
        });
      })
      .fail(() => {
        this.setState({error: err});
      });
  },
  render() {
    if (this.state.latest) {    
      return (<ReadingsTable readings={this.state.latest} nav={true} />)
    } else {
      return (
        <div>
          <h2> No Readings Found</h2>
          <p> Maybe you should add some?</p>
        </div>
      );
    }
  }
});

var Main = React.createClass({
  getInitialState() {
    return {};
  },
  componentDidMount() {
    $.get('/api/user', null, 'json')
      .done(data => {
        browserHistory.push('latest');
      })
      .fail(() => {
        this.setState({loggedin: false});
      });
  },
  render() {
      if (this.state.loggedin === false) {
        return (
          <div>
            <h2>Welcome</h2>
            <a href="api/authorize" className="btn btn-default btn-lg">Login</a>
            <a href="api/demo" className="btn btn-default btn-lg">Demo</a>
          </div>)
      }
      else {
        return (<img src="img/loading.svg" />)
      }
  }
});

render((
  <Router history={browserHistory} locales={['en-us']}>
    <Route path="/" component={Main} />
    <Route path="latest" component={Latest} />
    <Route path="readings/:coreid" component={ReadingsChart} />
  </Router>
), document.getElementById("content"));
