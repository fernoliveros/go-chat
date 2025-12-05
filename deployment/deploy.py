

def main():
    git_pull_latest()
    build_docker_image()
    run_docker_container()

def git_pull_latest():
    print("git pull from gochat repo")

def build_docker_image():
    print("docker buildx")

def run_docker_container():
    print("docker run container")

main()
# https://taskfile.dev/