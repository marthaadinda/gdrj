"use strict";

viewModel.chartComparison = new Object();
var cc = viewModel.chartComparison;

cc.dataTemp = [{ "activity": "drilling", "actual": 100, "plan": 80 }, { "activity": "completion", "actual": 85, "plan": 90 }, { "activity": "abandonment", "actual": 90, "plan": 85 }];
cc.data = {};
cc.ideas = ko.observableArray([]);
cc.getIdeas = function () {
	app.ajaxPost('/report/getdataanalysisidea', {}, function (res) {
		if (!app.isFine(res)) {
			return;
		}

		res.data.forEach(function (d) {
			if (cc.hasOwnProperty(d._id)) {
				cc.data[d._id](res.data);
			} else {
				cc.data[d._id] = ko.observableArray([]);
			}
		});

		cc.ideas(res.data);
		cc.selectedIdeas(cc.ideas().slice(0, 2));
		cc.render();
	});
};
cc.selectedIdeas = ko.observableArray([]);
cc.render = function () {
	var $container = $('.chart-container');
	$container.empty();

	cc.selectedIdeas().forEach(function (d) {
		var o = $("<div class=\"col-md-12 col-sm-12 no-padding hardcore\">\n\t\t\t<div class=\"chart chart-" + d._id + "\" style=\"height: 300px;\"></div>\n\t\t</div>");
		$container.append(o);

		var series = [{ field: 'actual' }, { field: 'plan' }];
		var data = cc.dataTemp;
		crt.createChart(".chart-" + d._id, d.name, series, data, 'activity');
	});
};

rpt.init = function () {
	cc.getIdeas();
};
rpt.refresh = function () {
	cc.render();
};