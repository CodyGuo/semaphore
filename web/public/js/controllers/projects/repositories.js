define(function () {
	app.registerController('ProjectRepositoriesCtrl', ['$scope', '$http', 'Project', '$uibModal', '$rootScope', 'SweetAlert', function ($scope, $http, Project, $modal, $rootScope, SweetAlert) {
		$scope.reload = function () {
		    $http.get(Project.getURL() + '/users?sort=name&order=asc').then(function (response) {
			  var users = response.data;
				$scope.project_user = null;
				$scope.users = users;

				for (var i = 0; i < users.length; i++) {
					if (users[i].id == $scope.user.id) {
						$scope.project_user = users[i];
						break;
					}
				}
			});
			$http.get(Project.getURL() + '/keys?type=ssh&sort=name&order=asc').then(function (keys) {
				$scope.sshKeys = keys.data;

				$http.get(Project.getURL() + '/repositories?sort=name&order=asc').then(function (repos) {
					repos.data.forEach(function (repo) {
						for (var i = 0; i < keys.data.length; i++) {
							if (repo.ssh_key_id == keys.data[i].id) {
								repo.ssh_key = keys.data[i];
								break;
							}
						}
					});

					$scope.repositories = repos.data;
				});
			});
		}

		$scope.remove = function (repo) {
			$http.delete(Project.getURL() + '/repositories/' + repo.id)
				.then(function () {
					$scope.reload();
				})
				.catch(function (response) {
					var d = response.data;
					if (!(d && d.templatesUse)) {
						SweetAlert.swal('error', 'could not delete repository..', 'error');
						return;
					}

					SweetAlert.swal({
						title: 'Repository in use',
						text: d.error,
						icon: 'error',
						buttons: {
							cancel: true,
							confirm: {
								text: 'Mark as removed',
								closeModel: false,
								className: 'bg-danger',
							}
						}
					}).then(function (value) {
						if (!value) {
							return;
						}

						$http.delete(Project.getURL() + '/repositories/' + repo.id + '?setRemoved=1')
							.then(function () {
								SweetAlert.stopLoading();
								SweetAlert.close();

								$scope.reload();
							})
							.catch(function () {
								SweetAlert.swal('Error', 'Could not delete repository..', 'error');
							});
					});
				});
		}

		$scope.update = function (repo) {
			var scope = $rootScope.$new();
			scope.keys = $scope.sshKeys;
			scope.repo = JSON.parse(JSON.stringify(repo));
            scope.project_user = $scope.project_user;

			$modal.open({
				templateUrl: '/tpl/projects/repositories/add.html',
				scope: scope
			}).result.then(function (opts) {
				if (opts.remove) {
					return $scope.remove(repo);
				}

				$http.put(Project.getURL() + '/repositories/' + repo.id, opts.repo).then(function () {
					$scope.reload();
				}).catch(function (response) {
					SweetAlert.swal('Error', 'Repository not updated: ' + response.status, 'error');
				});
			}, function () {
			});
		}

		$scope.add = function () {
			var scope = $rootScope.$new();
			scope.keys = $scope.sshKeys;
            scope.project_user = $scope.project_user;

			$modal.open({
				templateUrl: '/tpl/projects/repositories/add.html',
				scope: scope
			}).result.then(function (repo) {
				$http.post(Project.getURL() + '/repositories', repo.repo)
					.then(function () {
						$scope.reload();
					}).catch(function (response) {
					SweetAlert.swal('Error', 'Repository not added: ' + response.status, 'error');
				});
			}, function () {
			});
		}

		$scope.reload();
	}]);
});
