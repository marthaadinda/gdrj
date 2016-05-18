'use strict';

var vm = viewModel;

vm.pageTitle = ko.observable('Dashboard');
vm.menu = ko.observableArray([{ title: 'Dashboard', icon: 'home', href: '#', submenu: [] }, { title: 'Data', icon: 'user', href: '#', submenu: [{ title: 'Data Browser', href: '#', submenu: [] }, { title: 'Data Serializer', href: '#', submenu: [] }, { title: 'Web Grabber', href: '#', submenu: [] }, { title: 'File Browser', href: '#', submenu: [] }] }]);

vm.menuIcon = function (data) {
	return ko.computed(function () {
		return 'fa fa-' + data.icon;
	});
};

vm.prepareDropDownMenu = function () {
	$('ul.nav li.dropdown').hover(function () {
		$(this).find('.dropdown-menu').stop(true, true).fadeIn(200);
	}, function () {
		$(this).find('.dropdown-menu').stop(true, true).fadeOut(200);
	});
};

vm.prepareFilterToggle = function () {
	$('.material-switch input[type="checkbox"]').on('change', function () {
		var show = $(this).is(':checked');
		var $target = $(this).closest(".panel").find(".panel-filter");
		if (show) {
			$target.show(200);
		} else {
			$target.hide(200);
		}
	}).trigger('click');
};

$(function () {
	vm.prepareDropDownMenu();
	vm.prepareFilterToggle();
});