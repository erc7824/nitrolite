export const chainImageURLById = (id?: number) => {
    switch (id) {
        case 1:
        case 11155111:
            return "https://s2.coinmarketcap.com/static/img/coins/64x64/1027.png";
        case 137:
        case 80002:
            return "https://s2.coinmarketcap.com/static/img/coins/64x64/3890.png";
        case 59144:
        case 59141:
            return "https://s2.coinmarketcap.com/static/img/coins/64x64/27657.png";
        case 8453:
            return "https://assets.coingecko.com/asset_platforms/images/131/standard/base-network.png?1720533039";
        case 56:
            return "https://s2.coinmarketcap.com/static/img/coins/64x64/1839.png";
        case 100:
            return "https://s2.coinmarketcap.com/static/img/coins/64x64/1659.png";
        case 10:
            return "https://s2.coinmarketcap.com/static/img/coins/64x64/11840.png";
        case 534352:
            return "https://assets.coingecko.com/asset_platforms/images/153/standard/scroll.jpeg?1706606782";
        case 42161:
            return "https://s2.coinmarketcap.com/static/img/coins/64x64/11841.png";
        case 43114:
            return "https://s2.coinmarketcap.com/static/img/coins/64x64/5805.png";
        default:
            return null;
    }
};
