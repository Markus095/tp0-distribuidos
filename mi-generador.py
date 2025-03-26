import sys

def main(argc=len(sys.argv), argv=sys.argv):
    filename = argv[1]
    number_of_clients = int(argv[2])
    if number_of_clients < 1 :
        print("Number of clients must be greater than 0.")
        return
    try:
        with open(filename, 'w') as file:
            
            add_server_to_yaml(filename)
            for i in range(1, number_of_clients + 1):
                add_client_to_yaml(filename, i)
            add_network_to_yaml(filename)    
        print(f"YAML file '{filename}' created successfully.")
    except Exception as e:
        print(f"An error occurred: {e}")


def add_server_to_yaml(filename):
    try:
        with open(filename, 'a') as file:
            file.write("name: tp0\n")
            file.write("services:\n")
            file.write("  server:\n")
            file.write("    container_name: server\n")
            file.write("    image: server:latest\n")
            file.write("    entrypoint: python3 /main.py\n")
            file.write("    volumes:\n")
            file.write("      - ./server/config.ini:/config.ini\n")
            file.write("    environment:\n")
            file.write("      - PYTHONUNBUFFERED=1\n")
            file.write("    networks:\n")
            file.write("      - testing_net\n\n")
    except Exception as e:
        print(f"An error occurred: {e}")
    


def add_client_to_yaml(filename, client_number):
    try:
        with open(filename, 'a') as file:
            file.write(f"  client{client_number}:\n")
            file.write(f"    container_name: client{client_number}\n")
            file.write("    image: client:latest\n")
            file.write("    entrypoint: /client\n")
            file.write("    volumes:\n")
            file.write("      - ./client/config.yaml:/config.yaml\n")
            file.write(f"      - ././data/agency-{client_number}.csv:/dataset-{client_number}.csv\n")
            file.write("    environment:\n")
            file.write(f"      - CLI_ID={client_number}\n")
            file.write("    networks:\n")
            file.write("      - testing_net\n")
            file.write("    depends_on:\n")
            file.write("      - server\n\n")
    except Exception as e:
        print(f"An error occurred: {e}")

def add_network_to_yaml(filename):
    try:
        with open(filename, 'a') as file:
            file.write("networks:\n")
            file.write("  testing_net:\n")
            file.write("    ipam:\n")
            file.write("      driver: default\n")
            file.write("      config:\n")
            file.write("        - subnet: 172.25.125.0/24\n")
    except Exception as e:
        print(f"An error occurred: {e}")
    
main()