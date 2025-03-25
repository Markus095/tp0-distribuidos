import socket
import logging
import signal
from common.utils import Bet, store_bets

# Fix header and bet sizes
MessageHeaderSize = 6  # 4 bytes agencyNumber + 2 bytes num_bets
BetSize = 170  # 64 + 64 + 32 + 8 + 2 (matches client)

STORAGE_FILEPATH = "./bets.csv"

class Server:
    def __init__(self, port, listen_backlog):
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
            # Read the header (6 bytes)
            header = client_sock.recv(MessageHeaderSize)
            if len(header) < MessageHeaderSize:
                raise ValueError("Incomplete header received")

            # Extract agency ID (4 bytes) and number of bets (2 bytes)
            agency_id = int.from_bytes(header[0:4], byteorder='big')
            num_bets = int.from_bytes(header[4:6], byteorder='big')

            logging.debug(f"Received header: agency_id={agency_id}, num_bets={num_bets}")

            total_size = num_bets * BetSize
            msg = b""

            # Read the bets
            while len(msg) < total_size:
                chunk = client_sock.recv(min(1024, total_size - len(msg)))
                if not chunk:
                    raise ValueError("Connection closed before full message was received")
                msg += chunk

            # Decode and store bets
            bets = self.decode_bets(agency_id, num_bets, msg)
            store_bets(bets)

            # Log stored bets
            for bet in bets:
                logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | number: {bet.number}')

            # Send acknowledgment
            client_sock.sendall(b"OK")

        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def run(self):
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
        except Exception as e:
            logging.error(f"action: server_loop | result: error | error: {e}")
        finally:
            self.shutdown()
            logging.info("action: server_shutdown | result: success")

    def decode_bets(self, agency_id, num_bets, message: bytes) -> list[Bet]:
        """
        Decode a binary message into bets.
        """
        bets = []
        offset = 0

        for _ in range(num_bets):
            # Read FirstName (64 bytes)
            first_name = message[offset:offset+64].split(b'\0', 1)[0].decode('utf-8').strip()
            offset += 64

            # Read LastName (64 bytes)
            last_name = message[offset:offset+64].split(b'\0', 1)[0].decode('utf-8').strip()
            offset += 64

            # Read Document (32 bytes)
            document = message[offset:offset+32].split(b'\0', 1)[0].decode('utf-8').strip()
            offset += 32

            # Read birthdate (8 bytes) and convert to YYYY-MM-DD format
            date_str = message[offset:offset+8].decode('utf-8')
            birthdate = f"{date_str[:4]}-{date_str[4:6]}-{date_str[6:8]}"
            offset += 8

            # Read number (2 bytes) instead of 8
            number = int.from_bytes(message[offset:offset+2], byteorder='big')
            offset += 2  # Updated to match client

            # Create Bet object
            bet = Bet(str(agency_id), first_name, last_name, document, birthdate, str(number))
            bets.append(bet)

        return bets
