from common.utils import Bet, has_won, load_bets, store_bets

def process_bets(agency_id, num_bets, bets_data):
    decoded_bets = decode_bets(agency_id, num_bets, bets_data)
    store_bets(decoded_bets)
    return True

def decode_bets(agency_id, num_bets, message: bytes) -> list[Bet]:
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

def obtain_winners_documents():
    bets = load_bets()
    winners = [bet.document for bet in bets if has_won(bet)]
    return winners