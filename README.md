# CHEEZBURGER SCRAPER

The CheezBurger scraper is a Golang based application that provides an API interface to access and download the memes from https://icanhas.cheezburger.com/ in order.

## Project setup

```bash
 sudo docker build . -t cheezburger_scrapper
 sudo docker run --network host -e DEBUG=true -it cheezburger_scrapper
```

## Project structure
```
cheezburger_scraper/
    configs/
    controllers/
    downloads/ # Directory where to save downloaded files
    lib/ # Common useful packages
    middlewares/
    routes/ 
    test/ # Resources used for test
    types/ # Common types used along the app
```

## Configuration variables

- `(optional) PORT`               = Port where server runs
- `(optional) SITE_URL`           = Site url to scrap from
- `(optional) CARD_IMG_SELECTOR`  = Selector to get img components
- `(optional) MIN_CARDS_PER_PAGE` = Minimum cards per page in site. Used to parallelize processing
- `(optional) TIMEOUT`            = Timeout supported for each request
- `(optional) DEBUG`              = Debug option. If is set to `true` will display informative logs about the processing
- `(optional) DOWNLOADS_SAVE_DIR` = Directory where to save the downloaded images
- `(optional) SLEEP_TIME`         = Sleep time to await for resources


## Endpoints
* URL:
    `/images/downloads`
* Method:

    `GET`
* Query params:
    
    * `ammount`: number of images to download


    * `threads`: number of concurrent processes used to scrap the data

* Success Response:
    
    * **Code:** 200
    * **Content:** [`<url_of_image_1>`,`<url_of_image_2>`,...]

## Decisions taken
- I decided to implement an API structure to this project, since I understood in the interviews, that this is usually the work format used within propper. Having services that can retrive information, or act on third party pages, and from there grouping everything in an internal page.

- I decided to implement the parallelization of the processing only in the method that scouts the urls from the images, since this was the one that consumed most of the processing. It would be nice to implement it also in the section that downloads the images, but for now it didn't seem necessary.

- I decided to implement the Logger and Semaphore classes since this was the fastest, and most functional option for the moment. In a productive code I would take a better look at what libraries are already available to use, that fulfill the desired functionalities.