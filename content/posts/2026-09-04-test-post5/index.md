---
title: Test Post5
author: Ed The Hell
date: '2026-09-04'
slug: []
categories: []
tags: ["posts"]
---


``` r
library("ggplot2")
summary(penguins)
```

```
##       species          island       bill_len        bill_dep    
##  Adelie   :152   Biscoe   :168   Min.   :32.10   Min.   :13.10  
##  Chinstrap: 68   Dream    :124   1st Qu.:39.23   1st Qu.:15.60  
##  Gentoo   :124   Torgersen: 52   Median :44.45   Median :17.30  
##                                  Mean   :43.92   Mean   :17.15  
##                                  3rd Qu.:48.50   3rd Qu.:18.70  
##                                  Max.   :59.60   Max.   :21.50  
##                                  NAs    :2       NAs    :2      
##   flipper_len      body_mass        sex           year     
##  Min.   :172.0   Min.   :2700   female:165   Min.   :2007  
##  1st Qu.:190.0   1st Qu.:3550   male  :168   1st Qu.:2007  
##  Median :197.0   Median :4050   NAs   : 11   Median :2008  
##  Mean   :200.9   Mean   :4202                Mean   :2008  
##  3rd Qu.:213.0   3rd Qu.:4750                3rd Qu.:2009  
##  Max.   :231.0   Max.   :6300                Max.   :2009  
##  NAs    :2       NAs    :2
```

``` r
ggplot(data=penguins) +
    geom_point(aes(x=bill_len, y=bill_dep, color=species))
```

```
## Warning: Removed 2 rows containing missing values or values outside the scale range
## (`geom_point()`).
```

<img src="{{< blogdown/postref >}}index_files/figure-html/chunk_one-1.png" alt="" width="672" />
