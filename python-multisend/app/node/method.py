# keccak256('approve(address,uint256)')+hex(address)+hex(uint256)
def approve(to, value, decimals):
    value = int(value * decimals)
    return "0x095ea7b3{:064x}".format(int(to, 0)) + "{:064x}".format(2 ** 256 - 1 if value < 0 else value)


# keccak256('allowance(address,address)')+hex(address)+hex(address)
# def allowance(owner, spender):
#     return "0xdd62ed3e{:064x}".format(int(owner, 0)) + "{:064x}".format(int(spender, 0))

#"multisendsToken(address,address[],address[],uint256[])": "be5c23d3"
def multisend(tokenAddress, spender, receiver, value, decimal):
    count = len(spender)
    if len(value) != count:
        return {"error": "address and values not matched"}
    t = ""
    offset = 0 + 32*5
    t += "{:064x}".format(offset)

    offset = offset + (count + 1) * 32
    t += "{:064x}".format(offset)

    offset = offset + (count + 1) * 32
    t += "{:064x}".format(offset)

    offset = offset + (count + 1) * 32
    t += "{:064x}".format(offset)

    t += "{:064x}".format(count)
    for i in range(count):
        t += "{:064x}".format(int(spender[i], 0))

    t += "{:064x}".format(count)
    for i in range(count):
        t += "{:064x}".format(int(receiver[i], 0))

    t += "{:064x}".format(count)
    for i in range(count):
        t += "{:064x}".format(value[i] * decimal)

    return "0xbe5c23d3{:064x}".format(int(tokenAddress, 0)) + t

