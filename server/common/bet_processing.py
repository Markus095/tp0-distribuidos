from common.utils import Bet, has_won, load_bets, store_bets

import logging
NAME_SIZE = 64
SURNAMES_SIZE = 64
DOCUMENT_SIZE = 8
BIRTHDATE_SIZE = 8
CHOSEN_NUMBER_SIZE = 2


def process_bets(agency_id, num_bets, bets_data, lock):
    decoded_bets = decode_bets(agency_id, num_bets, bets_data)
    with lock:
        store_bets(decoded_bets)
    return True

def decode_bets(agency_id, num_bets, message: bytes) -> list[Bet]:
    """
    Decode a binary message into bets.
    """
    bets = []
    offset = 0

    for _ in range(num_bets):
        first_name = message[offset:offset+NAME_SIZE].split(b'\0', 1)[0].decode('utf-8').strip()
        offset += NAME_SIZE

        last_name = message[offset:offset+SURNAMES_SIZE].split(b'\0', 1)[0].decode('utf-8').strip()
        offset += SURNAMES_SIZE

        document = message[offset:offset+DOCUMENT_SIZE].split(b'\0', 1)[0].decode('utf-8').strip()
        offset += DOCUMENT_SIZE

        date_str = message[offset:offset+BIRTHDATE_SIZE].decode('utf-8')
        birthdate = f"{date_str[:4]}-{date_str[4:6]}-{date_str[6:8]}"
        offset += BIRTHDATE_SIZE

        number = int.from_bytes(message[offset:offset+CHOSEN_NUMBER_SIZE], byteorder='big')
        offset += CHOSEN_NUMBER_SIZE

        bet = Bet(str(agency_id), first_name, last_name, document, birthdate, str(number))
        bets.append(bet)

    return bets

def obtain_winners_documents():
    bets = list(load_bets())
    winners = [(bet.agency, bet.document) for bet in bets if has_won(bet)]
    return winners