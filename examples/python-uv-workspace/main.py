import workspace_package

def main():
    # test.json doesn't like newlines
    print("Hello from python-uv-workspace!", end="")
    workspace_package.main()


if __name__ == "__main__":
    main()
