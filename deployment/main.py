import docker
from git import Repo
import os
import time

gochat_repo_path = '/home/fernoliveros/dev/go-chat'
gochat_docker_image_tag = 'gochat'
gochat_docker_container_name = 'gochatc'
git_remote_name = 'origin'
git_branch = 'main'


def check_remote_updates():
    repo = Repo(gochat_repo_path)
    origin = repo.remote(git_remote_name)
    origin.fetch() # Fetch all the new information about the remote repository

    # Compare local HEAD with the remote tracking branch
    local_commit = repo.head.commit
    remote_commit = origin.refs[git_branch].commit

    if local_commit != remote_commit:
        print(f"Remote {git_branch} has updates not in local. Local commit: {local_commit}, Remote commit: {remote_commit}")
        git_pull_latest()
        docker_stop_build_and_deploy()
    else:
        print(f"Local branch {git_branch} is up-to-date with remote.")

def docker_stop_build_and_deploy():
    docker_client = docker.from_env()

    stop_docker_container(docker_client)
    prune_docker_images(docker_client)
    build_docker_image(docker_client)
    run_docker_container(docker_client)

def git_pull_latest():

    if os.path.exists(gochat_repo_path):
        # Open the repository object
        repo = Repo(gochat_repo_path)
        assert not repo.bare

        print(f"Repository {gochat_repo_path} is open.")
        print(f"Current active branch: {repo.active_branch.name}")
        print(f"Is the working tree dirty? {repo.is_dirty()}")
    else:
        print(f"Repository path not found at {gochat_repo_path}")

def build_docker_image(docker_client):
    
    print("docker build image")
    image, build_logs = docker_client.images.build(path=gochat_repo_path, tag=gochat_docker_image_tag)

    print(f"Successfully built image: {image.tags[0]}")

    # Optional: Stream and print build logs
    print("\n--- Build Logs ---")
    for log in build_logs:
        if 'stream' in log:
            print(log['stream'].strip())
        elif 'error' in log:
            print(f"Error: {log['error']}")


def run_docker_container(docker_client):
    print("run docker container")
    port_mapping = {"8080/tcp": "8080"}

    try:
        # Run the container in detached mode (-d) with the specified port mapping
        container = docker_client.containers.run(
            image=gochat_docker_image_tag,
            detach=True,
            ports=port_mapping,
            name=gochat_docker_container_name
        )

        print(f"Container '{container.name}' started successfully.")
        print(f"Mapped container port 80/tcp to host port 8080.")
        print(f"Access the running container at http://localhost:8080")

    except docker.errors.ImageNotFound:
        print(f"Error: The '{gochat_docker_image_tag}' image was not found locally. Pulling it might be necessary.")
    except docker.errors.ContainerError as e:
        print(f"Error running container: {e}")
    except Exception as e:
        print(f"An unexpected error occurred: {e}")

def stop_docker_container(docker_client):
    try:
        container = docker_client.containers.get(gochat_docker_container_name)
        
        print(f"Stopping container: {container.name} ({container.id})")
        container.stop()
        
        print(f"Container {container.name} stopped.")

    except docker.errors.NotFound:
        print(f"Container {gochat_docker_container_name} not found.")
    except Exception as e:
        print(f"An error occurred: {e}")


def prune_docker_images(docker_client):
    filters = {'dangling': True}
    
    result = docker_client.images.prune(filters=filters)
    
    print(f"Removed image IDs: {result['ImagesDeleted']}")
    print(f"Total space reclaimed: {result['SpaceReclaimed']} bytes")

if __name__ == "__main__":
    print("Starting remote git branch monitor (polling every 60 seconds)...")
    while True:
        check_remote_updates()
        time.sleep(60)
