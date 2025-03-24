import socket
import logging
import signal
from utils import Bet, store_bets 

HEADER_SIZE = 8  # Constant value
BET_SIZE = 172  # Constant value

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def __handle_sigterm(self, signum, frame):
        logging.info("action: handle_signal | result: success")
        self._running = False

    def shutdown(self):
        logging.info("action: shutdown_server | result: in_progress")
        
        if self._server_socket:
            try:
                self._server_socket.close()
                logging.info("action: close_server_socket | result: success")
            except Exception as e:
                logging.error(f"action: close_server_socket | result: fail | error: {e}")
        
        logging.info("action: shutdown_server | result: success")
    



    def __handle_client_connection(self, client_sock):
        """
        Handle incoming bet registration from clients
        """
        try:
            # Read the header (first 8 bytes)
            header = client_sock.recv(HEADER_SIZE)
            if len(header) < HEADER_SIZE:
                raise ValueError("Incomplete header received")
            
            # Extract ID and number of chunks from the header
            client_id = header[0]
            num_chunks = int.from_bytes(header[1:HEADER_SIZE], byteorder='big')
            
            # Calculate the total message size
            total_size = num_chunks * BET_SIZE
            
            # Read the rest of the message in a loop
            msg = b""
            while len(msg) < total_size:
                chunk = client_sock.recv(min(1024, total_size - len(msg)))
                if not chunk:
                    raise ValueError("Connection closed before full message was received")
                msg += chunk
            
            # Decode and store bets
            bets = self.decode_bets(msg)
            store_bets(bets)
            
            # Log each bet stored
            for bet in bets:
                logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')
            
            # Send acknowledgment
            client_sock.sendall(b"OK")
            
        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        try:
            while self._running:
                try:
                    self._server_socket.settimeout(1.0)
                    client_sock = self.__accept_new_connection()
                    if client_sock:
                        self.__handle_client_connection(client_sock)
                except socket.timeout:
                    continue
                except Exception as e:
                    if self._running:
                        logging.error(f"action: server_loop | result: error | error: {e}")
        finally:
            self.shutdown()

    def decode_bets(self, message: bytes) -> list[Bet]:
        """
        Decode a binary message into client ID and list of bets.
        Message format:
        - 1 byte: client ID (agency)
        - 7 bytes: number of bets
        - For each bet (172 bytes):
            - 64 bytes: first name (null padded)
            - 64 bytes: last name (null padded)
            - 32 bytes: document (null padded)
            - 8 bytes: birthdate (YYYYMMDD)
            - 2 bytes: number
        """
        # Read client ID/agency (first byte)
        agency_id = message[0]
        
        # Read number of bets (next 7 bytes)
        num_bets = int.from_bytes(message[1:8], byteorder='big')
        
        bets = []
        offset = 8  # Start after header
        
        for _ in range(num_bets):
            # Read FirstName (64 bytes)
            first_name = message[offset:offset+64].split(b'\0', 1)[0].decode('utf-8')
            offset += 64
            
            # Read LastName (64 bytes)
            last_name = message[offset:offset+64].split(b'\0', 1)[0].decode('utf-8')
            offset += 64
            
            # Read Document (32 bytes)
            document = message[offset:offset+32].split(b'\0', 1)[0].decode('utf-8')
            offset += 32
            
            # Read birthdate (8 bytes) and convert to YYYY-MM-DD format
            date_str = message[offset:offset+8].decode('utf-8')
            birthdate = f"{date_str[:4]}-{date_str[4:6]}-{date_str[6:8]}"
            offset += 8
            
            # Read number (2 bytes)
            number = int.from_bytes(message[offset:offset+2], byteorder='big')
            offset += 2
            
            bet = Bet(str(agency_id), first_name, last_name, document, birthdate, str(number))
            bets.append(bet)
        
        return bets
    
