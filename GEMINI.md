# Project Configuration

## Environment
- Cloud Shell Home: `/h`
- Project Root: `/h/image_rotation_correction`

## Container Environment (`gocv`)
- Container Name: `gocv`
- Base Image: `gocv/opencv` (Required for OpenCV API version compatibility)
- Working Directory Mapping: The project root is accessible via the container's mounted path corresponding to `/h/image_rotation_correction`.

## Command Execution
To execute commands within the container, use:
`docker exec gocv <cmd>`
(Prefix with `!` if running from within the CLI environment if required by your tooling, otherwise just run the command).
