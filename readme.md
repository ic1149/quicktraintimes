# Quick Train Times

## about

Quick train times (QTT) is an application that gets live train departure data in the UK (National Rail) instantly.
Traditionally, it is required for user to choose the origin and destination stations every time.
Everyday going to and from school or work, we have to select the stations manually.
With QTT, entries with time frames can be set to automatically show your routine train journeys.
All of the critical information is presented in one view.

## how to install

### linux
Run the install script `install.sh` if you have a desktop environment, this installs the program for all users, with a desktop file.

Alternatively, downoad quicktraintimes.tar.xz in release. Then, extract the file. After that, cd into it and run `sudo make install`. The command `quicktraintimes` runs the app.

### android
- download the apk file in release
- install the apk file (you may have to enable permission to install apk)

### windows
- download the exe file in release
- run the file

## how to use the app

### 1. Key

Go to the Settings page.
Set your LDBWS Departure API Key here.
You can also set other preferences.
Remember to save the options.
Restart the app to apply the settings.

A key can be obtained by subscribing [Live Deaprture Board on Rail Data Marketplace](https://raildata.org.uk/dataProduct/P-d81d6eaf-8060-4467-a339-1c833e50cbbe/overview)


### 2. Config QTTs

Go to the Config QTTs page. Create new entries here. Fill in the required parameters. Remember to click save for each entry. Go back to homepage and your train times will appear if within the desired time slots. Please note if more than two entries are within current time, only the first two will show. 

### example QTT entries

#### example 1 - a commuter living in London, working in Bristol

I am going to work on weekdays. Usually, I wake up 06:30 and take a train around 08:00 at London Paddington to Bristol Temple Meads.
```
Start time 06:30
End time 11:00
From station PAD
To station BRI
Days Mon Tue Wed Thu Fri
```
After work, I arrive at Bristol Temple Meads around 18:30 to take the train back home.
```
Start time 18:30
End time 20:00
From station BRI
To station PAD
Days Mon Tue Wed Thu Fri
```

On some Saturdays, I visit my friend living in Maidenhead. However, the time varies. I may even stay overnight sometime.
```
Start time 00:00
End time 23:59
From station PAD
To station MAI
Days Sat

Start time 00:00
End time 23:59
From station MAI
To station PAD
Days Sat Sun
```
