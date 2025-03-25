import socket
import logging
import signal
from utils import Bet, store_bets 

MessageHeaderSize = 20  # Updated to match client's protocol
BetSize = 176  # Updated to match client's protocol

STORAGE_FILEPATH = "./bets.csv"

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
            # Read the header (20 bytes)
            header = client_sock.recv(MessageHeaderSize)
            if len(header) < MessageHeaderSize:
                raise ValueError("Incomplete header received")
            
            # Extract agency ID (first 4 bytes) and number of bets (next 16 bytes)
            agency_id = int.from_bytes(header[0:4], byteorder='big')
            num_bets = int.from_bytes(header[4:20], byteorder='big')
            
            # Calculate the total message size (BetSize = 176 bytes per bet)
            total_size = num_bets * BetSize
            
            # Read the rest of the message in a loop
            msg = b""
            while len(msg) < total_size:
                chunk = client_sock.recv(min(1024, total_size - len(msg)))
                if not chunk:
                    raise ValueError("Connection closed before full message was received")
                msg += chunk
            
            # Decode and store bets
            bets = self.decode_bets(header, msg)
            logging.debug(f"Attempting to store bets at: {STORAGE_FILEPATH}")
            store_bets(bets)
            
            # Verify if file exists and its contents
            try:
                with open(STORAGE_FILEPATH, 'r') as f:
                    logging.debug(f"File contents after store:\n{f.read()}")
            except Exception as e:
                logging.error(f"Failed to read file after storage: {e}")
            
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
        Server loop that accepts new connections and handles client communication
        """
        logging.info("action: server_start | result: in_progress")
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
                        # Don't exit on error, try to continue running
        except Exception as e:
            logging.error(f"action: server_loop | result: error | error: {e}")
        finally:
            self.shutdown()
            logging.info("action: server_start | result: success")

    def decode_bets(self, header: bytes, message: bytes) -> list[Bet]:
        """
        Decode a binary message into client ID and list of bets.
        Message format:
        Header (20 bytes):
            - 4 bytes: agency ID
            - 16 bytes: number of bets
        For each bet (176 bytes):
            - 64 bytes: first name (null padded)
            - 64 bytes: last name (null padded)
            - 32 bytes: document (null padded)
            - 8 bytes: birthdate (YYYYMMDD)
            - 8 bytes: number
        """
        # Read agency ID from header (first 4 bytes)
        agency_id = int.from_bytes(header[0:4], byteorder='big')
        
        # Read number of bets from header (next 16 bytes)
        num_bets = int.from_bytes(header[4:20], byteorder='big')
        
        bets = []
        offset = 0
        
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
            
            # Read number (8 bytes)
            number = int.from_bytes(message[offset:offset+8], byteorder='big')
            offset += 8
            
            bet = Bet(str(agency_id), first_name, last_name, document, birthdate, str(number))
            bets.append(bet)
        
        return bets
    
