# Image to JSON

This service exposes an endpoint which takes an image payload and returns json containing all the alcoholic beverages in the menu.
Each item in the returned list has the following fields:

| Field name | Description                                                                                                                                                                                  | Example                   |
| ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------- |
| Drink      | Name of the drink as it is found in the menu                                                                                                                                                 | Captain morgan, Heineken  |
| Type       | Category of beverage                                                                                                                                                                         | Beer, white wine, tequila |
| Price      | Price of the item, does not take currency into account as all menus will be assumed to use sek                                                                                               | 1337                      |
| Size       | This value comes either from right next to the name of the drink, ie "Carlsberg 50cl", or in the case of wines where you pay either by glass or by bottle. In most cases this field is empty | 50cl, glass, bottle       |
| Tap        | Wether or not the drink is served from tap or bottle/can                                                                                                                                     | true, false               |

## Usage

### Build

Create the docker image with `docker build -t imagetojson .`

Run the container with `docker run --env-file .env -p 8080:8080 imagetojson`

### Example call

curl -X POST http://localhost:8080/imagetojson -H "Content-Type: image/jpg" --data-binary "@menu2.png"

### Result

```[
{"drink": "Proverb Pinot Grigio", "type": "white wine", "price": 8, "size": "glass", "tap": false},
{"drink": "Proverb Pinot Grigio", "type": "white wine", "price": 32, "size": "bottle", "tap": false},
{"drink": "Proverb Chardonnay", "type": "white wine", "price": 8, "size": "glass", "tap": false},
{"drink": "Proverb Chardonnay", "type": "white wine", "price": 32, "size": "bottle", "tap": false},
{"drink": "Whitehaven Sauvignon Blanc", "type": "white wine", "price": 12, "size": "glass", "tap": false},
{"drink": "Whitehaven Sauvignon Blanc", "type": "white wine", "price": 40, "size": "bottle", "tap": false},
{"drink": "Archerry Summit Pinot Gris", "type": "white wine", "price": 12, "size": "glass", "tap": false},
{"drink": "Archerry Summit Pinot Gris", "type": "white wine", "price": 40, "size": "bottle", "tap": false},
{"drink": "Brown Estate Chardonnay", "type": "white wine", "price": 12, "size": "glass", "tap": false},
{"drink": "Brown Estate Chardonnay", "type": "white wine", "price": 40, "size": "bottle", "tap": false},
{"drink": "Simonet Brut Blanc de Blanc", "type": "sparkling wine", "price": 12, "size": "glass", "tap": false},
{"drink": "Simonet Brut Blanc de Blanc", "type": "sparkling wine", "price": 35, "size": "bottle", "tap": false},
{"drink": "Proverb Cabernet", "type": "red wine", "price": 8, "size": "glass", "tap": false},
{"drink": "Proverb Cabernet", "type": "red wine", "price": 32, "size": "bottle", "tap": false},
{"drink": "Proverb Merlot", "type": "red wine", "price": 8, "size": "glass", "tap": false},
{"drink": "Proverb Merlot", "type": "red wine", "price": 32, "size": "bottle", "tap": false},
{"drink": "Inscription Pinot Noir", "type": "red wine", "price": 14, "size": "glass", "tap": false},
{"drink": "Inscription Pinot Noir", "type": "red wine", "price": 40, "size": "bottle", "tap": false},
{"drink": "Architect Cabernet Sauvignon", "type": "red wine", "price": 15, "size": "glass", "tap": false},
{"drink": "Architect Cabernet Sauvignon", "type": "red wine", "price": 40, "size": "bottle", "tap": false},
{"drink": "Catena Malbec", "type": "red wine", "price": 15, "size": "glass", "tap": false},
{"drink": "Catena Malbec", "type": "red wine", "price": 40, "size": "bottle", "tap": false},
{"drink": "E Bourbon", "type": "whiskey", "price": 12, "size": "", "tap": false},
{"drink": "Jack Daniels", "type": "whiskey", "price": 12, "size": "", "tap": false},
{"drink": "Slane Irish Whiskey", "type": "whiskey", "price": 12, "size": "", "tap": false},
{"drink": "Killarney Irish Whiskey", "type": "whiskey", "price": 12, "size": "", "tap": false},
{"drink": "Old Forester", "type": "whiskey", "price": 12, "size": "", "tap": false},
{"drink": "Johnnie Walker Black", "type": "whiskey", "price": 14, "size": "", "tap": false},
{"drink": "Makers Mark", "type": "whiskey", "price": 15, "size": "", "tap": false},
{"drink": "Bulleit Bourbon", "type": "whiskey", "price": 16, "size": "", "tap": false},
{"drink": "Bulleit Rye", "type": "whiskey", "price": 16, "size": "", "tap": false},
{"drink": "Woodford", "type": "whiskey", "price": 16, "size": "", "tap": false},
{"drink": "Heaven\u2019s Door Rye", "type": "whiskey", "price": 16, "size": "", "tap": false},
{"drink": "Heaven\u2019s Door Straight Bourbon", "type": "whiskey", "price": 18, "size": "", "tap": false},
{"drink": "Heaven\u2019s Door Double Barrel", "type": "whiskey", "price": 18, "size": "", "tap": false},
{"drink": "E Gin Bombay", "type": "gin", "price": 10, "size": "", "tap": false},
{"drink": "Sapphire", "type": "gin", "price": 12, "size": "", "tap": false},
{"drink": "Tanqueray", "type": "gin", "price": 14, "size": "", "tap": false},
{"drink": "Hendricks", "type": "gin", "price": 16, "size": "", "tap": false},
{"drink": "E Vodka", "type": "vodka", "price": 10, "size": "", "tap": false},
{"drink": "Titos", "type": "vodka", "price": 12, "size": "", "tap": false},
{"drink": "Ketel One", "type": "vodka", "price": 14, "size": "", "tap": false},
{"drink": "Grey Goose", "type": "vodka", "price": 16, "size": "", "tap": false},
{"drink": "Milagro", "type": "tequila", "price": 12, "size": "", "tap": false},
{"drink": "Herradura Silver", "type": "tequila", "price": 14, "size": "", "tap": false},
{"drink": "Herradura Reposado", "type": "tequila", "price": 14, "size": "", "tap": false},
{"drink": "Don Julio", "type": "tequila", "price": 16, "size": "", "tap": false},
{"drink": "Casamigos", "type": "tequila", "price": 18, "size": "", "tap": false},
{"drink": "Bacardi", "type": "rum", "price": 12, "size": "", "tap": false},
{"drink": "Captain Morgan", "type": "rum", "price": 12, "size": "", "tap": false},
{"drink": "Appleton Estates", "type": "rum", "price": 14, "size": "", "tap": false}
]
```
