#!/bin/bash
echo "Validating YAML files..."
for package_directory in "$BASE_DIRECTORY"/*; do
  echo $package_directory
    # Check if it's a directory
    if [ -d "$package_directory" ]; then
          # Print the package name (directory name) and the content of collection.yaml
          if [ -f "$package_directory/collection.yaml" ]; then
              diff "$package_directory"/collection.yaml "$package_directory"/collection.yaml.golden

              if [ $? -ne 0 ]; then
                  echo "Changes don't match the golden file for package $package_directory. Exiting..."
                  exit 1
              fi
          fi

          if [ -f "$package_directory/definition.yaml" ]; then
              diff "$package_directory"/definition.yaml "$package_directory"/definition.yaml.golden

              if [ $? -ne 0 ]; then
                  echo "Changes don't match the golden file for package $package_directory. Exiting..."
                  exit 1
              fi
          fi
    fi
done