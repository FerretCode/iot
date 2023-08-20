load('ext://restart_process', 'docker_build_with_restart')

allow_k8s_contexts('iot')

docker_build_with_restart('sthanguy/iot-user',
							context='./services/user',
							entrypoint='go run main.go',
							dockerfile='./services/user/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/user', '/home/route'),
							]
)

"""
docker_build_with_restart('sthanguy/iot-aggregator',
							context='./services/aggregator',
							entrypoint='go run main.go',
							dockerfile='./services/aggregator/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/aggregator', '/home/route'),
							]
)
"""

docker_build_with_restart('sthanguy/iot-gateway',
							context='./services/gateway',
							entrypoint='go run main.go',
							dockerfile='./services/gateway/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/gateway', '/home/route'),
							]
)

"""
docker_build_with_restart('sthanguy/iot-teams',
							context='./services/teams',
							entrypoint='go run main.go',
							dockerfile='./services/teams/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/teams', '/home/route'),
							]
)
"""


k8s_yaml(['manifests/user/deployment.yml', 'manifests/user/service.yml'])
#k8s_yaml(['manifests/teams/deployment.yml', 'manifests/teams/service.yml'])
#k8s_yaml(['manifests/aggregator/deployment.yml', 'manifests/aggregator/service.yml'])
k8s_yaml(['manifests/gateway/deployment.yml', 'manifests/gateway/service.yml', 'manifests/gateway/ingress.yml'])
