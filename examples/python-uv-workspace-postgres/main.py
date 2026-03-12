from db_package import get_connection_string


def main():
    print(f"Connection: {get_connection_string()}", end="")


if __name__ == "__main__":
    main()
