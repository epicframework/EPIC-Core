$(document).ready(function() {

	$(".action-trigger").click(function() {

		var target = $(this).data("target");
		var action = $(this).data("action");

		data = {"target": target, "action": action}

		$.ajax({
			method: "POST",
			url: "/devices/action/trigger",
			data: data,
			success: function(data) {
				alert("Triggered "+action+" for "+target+": "+data);
			}
		})

	});

	$(".submit-read").click(function() {

		var button = $(this);
		var target = $(this).data('target');
		var property = $(this).data('property');

		data = {"target": target, "property": property}

		$.ajax({
			method: "POST",
			url: "/devices/property/read",
			data: {
				"target": target,
				"property": property,
			},
			success: function(data) {
				console.log(data);
				response = JSON.parse(data);
				console.log(response["value"]);
				button.siblings('.read-value-display').val(response["value"]);
			}
		})

	});

	$(".submit-write").click(function() {

		var target = $(this).data('target');
		var property = $(this).data('property');
		var value = $('input[name="'+$(this).data('input')+'"]').val();

		data = {"target": target, "property": property, "value": value}

		$.ajax({
			method: "POST",
			url: "/devices/property/write",
			success: function(data) {
				alert("Wrote "+property+" of "+target+" to "+value+": "+data)
			}
		})

	});

	$('#sidebarCollapse').click(function() {
		$('#sidebar').toggleClass('active');
	});
	
	$('#search-bar').on('change', function() {
		
	});

});