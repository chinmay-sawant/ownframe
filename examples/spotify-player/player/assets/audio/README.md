# Bundled soundtrack

The Spotify example ships 30 distinct tracks from [Open Lo-Fi v1.0.0](https://github.com/btahir/open-lofi/releases/tag/v1.0.0). The source project says its maintainer generated the music with Suno v5 and released it under [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/). The catalog records the source, audio license, lengths, checksums, categories, and player descriptions. The short descriptions are written for this demo.

The tracks are re-encoded as 96 kbps stereo MP3 at 32 kHz. The Go code remains under the repository's MIT license; the music has its own CC0 terms.

## Static cover art

These local 512 × 512 JPEGs are embedded with the player, so covers do not depend on a remote image API. The ten category images support recommendation cards, radio, library, and fallback artwork. The Popular list uses 30 individual track covers. All cover images are center-cropped to 512 × 512 and compressed for the app.

| Category | Image | Creator | License and source |
| --- | --- | --- | --- |
| Focus | `focus.jpg` | btahir / Open Lo-Fi | [CC0 1.0](https://github.com/btahir/open-lofi/blob/main/LICENSE), [image](https://github.com/btahir/open-lofi/blob/main/assets/opengraph.png) |
| Seasons | `spring.jpg` | Pandhe Wisnu | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [image](https://commons.wikimedia.org/wiki/File:Sakura_After_The_Rain_(206649015).jpeg) |
| Ambient | `ambient.jpg` | Michala Garrison / NASA Earth Observatory | Public domain, [image](https://commons.wikimedia.org/wiki/File:A_Colorful_Aurora_Paints_the_Night_Sky_(153463).jpg) |
| Soul | `soul.jpg` | Amitbalani | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [image](https://commons.wikimedia.org/wiki/File:HMV_Record_Player.jpg) |
| Late night | `late-night.jpg` | PattayaPatrol | [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/), [image](https://commons.wikimedia.org/wiki/File:DFC_5409_Neon_reflections_a_food_truck_glows_beside_a_rooftop_pool_casting_colorful_lights_across_the_water_in_Pattaya_City.jpg); crop shared under CC BY-SA 4.0 |
| Funk | `funk.jpg` | Amitbalani | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [image](https://commons.wikimedia.org/wiki/File:HMV_Stereo_1010_Gramophone_with_Speakers.jpg) |
| Jazz | `jazz.jpg` | Photographer's Mate 1st Class Ted Banks / U.S. Navy | Public domain, [image](https://commons.wikimedia.org/wiki/File:US_Navy_041215-N-3236B-009_Musician_2nd_Class_Jeremy_Bustillos,_a_saxophone_player_in_the_Commander_Sixth_Fleet_Jazz_Ensemble,_entertains_World_War_II_veterans,_active_duty_Soldiers_and_Sailors_during_a_luncheon_honoring_the_li.jpg) |
| Chillhop | `chillhop.jpg` | PattayaPatrol | [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/), [image](https://commons.wikimedia.org/wiki/File:DFC_2075_A_quiet_urban_street_at_dusk_lined_with_small_eateries_string_lights_overhead_and_cars_and_motorcycles_moving_along_the_road.jpg); crop shared under CC BY-SA 4.0 |
| Asian / Zen | `asian-zen.jpg` | Balon Greyjoy | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [image](https://commons.wikimedia.org/wiki/File:20190202_Nagasaki_Chinese_New_Year_lantern-1.jpg) |
| Hybrid | `hybrid.jpg` | pwbaker | [CC BY 2.0](https://creativecommons.org/licenses/by/2.0/), [image](https://commons.wikimedia.org/wiki/File:Coffee_house_AND_cafe.jpg) |

Audio and cover art have separate license terms; the artwork credits above include the attribution required for the listed Creative Commons licenses.

## Individual Popular covers

Each bundled track has a different static cover. The 30 additional photos were indexed by [Openverse](https://openverse.org/) as CC0 and are distributed here as 512 × 512 JPEGs. The catalog also records each Openverse ID, search query, and thumbnail URL.

| File | Source title | Creator | License and source |
| --- | --- | --- | --- |
| `track-01.jpg` | Books of Knowledge Picton Library Liverpool | Terry Kearney | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/24490288@N04/16711663295) |
| `track-02.jpg` | nature-flowers-garden-tree | pixellaphoto | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/137643065@N06/24298701426) |
| `track-03.jpg` | Reykjanes Geopark Aurora | Seanie2322 | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/114717511@N02/25941120808) |
| `track-04.jpg` | Concert Stage | Freestocks.org | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://stocksnap.io/photo/concert-stage-5GPP7A6LVW) |
| `track-05.jpg` | iGuitar | cogdogblog | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/37996646802@N01/4385871127) |
| `track-06.jpg` | Dusk approaches crowded entertainment district | Carol M Highsmith | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.rawpixel.com/image/8035873/photo-image-person-neon-road) |
| `track-07.jpg` | Vinyl Record | Skitter Photo | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://stocksnap.io/photo/vinyl-record-0NOS06CDVI) |
| `track-08.jpg` | Musician Playing Saxophone | Image Catalog | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/132795455@N08/20311658263) |
| `track-09.jpg` | Cafe interior from Hängmatta | Anna Wolf Art | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/48042762@N05/11117173633) |
| `track-10.jpg` | Patio table and chair close up, white contemporary Buddha statue, night lights, Japanese lantern, lotus fountain, pink roses, bamboo fence, A Garden for the Buddha, Seattle, Washington, USA | Wonderlane | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/71401718@N00/9436463291) |
| `track-11.jpg` | starry-night-sky | diamondtdesign | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/155872632@N02/39776804955) |
| `track-12.jpg` | Coffee Cup with Coffee on Red Background | Image Catalog | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/132795455@N08/17625638243) |
| `track-13.jpg` | Headphones Music | JESHOOTS.com | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://stocksnap.io/photo/headphones-music-UXKE7VWPCY) |
| `track-14.jpg` | Piano keyboard | freestocks.org | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/135396164@N05/22384280069) |
| `track-15.jpg` | background waterfalls magnificent | Ⓔⓛⓘⓢⓔ ❀❁ | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/128972682@N04/17290739282) |
| `track-16.jpg` | foggy forest | brenkee | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/55128416@N05/38568484206) |
| `track-17.jpg` | Rainy Day | romainguy | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/24046097@N00/11395179386) |
| `track-18.jpg` | RB0_0255 | Bence Kulcsar | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/47284666@N05/7825465930) |
| `track-19.jpg` | Record Player Limited Edition Pink Vinyl Aimee Mann Mental Illness | John K Thorne | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/89918055@N05/34053232855) |
| `track-20.jpg` | Snowy night city lights calm and lovely, Space Needle, Lake Union, Gas Works Park, Seattle, Washington, USA | Wonderlane | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/71401718@N00/6739756699) |
| `track-21.jpg` | Marshall Beach Sunset | romainguy | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/24046097@N00/6753666737) |
| `track-22.jpg` | Colorful Flower Garden | markus spiske | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/125167502@N02/48242898637) |
| `track-23.jpg` | Lighthouse Coast | Dave Meier | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://stocksnap.io/photo/lighthouse-coast-A189DFB220) |
| `track-24.jpg` | São Paulo - Skyline by night | Andre Deak | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/30578078@N00/1816831388) |
| `track-25.jpg` | Praktica BC1 - Parked Cars at Night | Kojotisko | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/98796230@N03/13149331513) |
| `track-26.jpg` | 2010.02 - 'Photo of a decorated Dutch train - rail-car at the train-station decorated with graffiti art, location at the Central Station in Amsterdam - geo-tagged free urban photo, in public domain / Commons; Fons Heijnsbroek, The Netherlands | Amsterdam free photos & pictures of the Dutch city | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/104736837@N03/27657675041) |
| `track-27.jpg` | Urban street-art and wall-painting art on the corner Anjelierstraat and the canal Prinsengracht on the border of Jordaan district Amsterdam, November 2013; photo Amsterdam city; urban photographer Fons Heijnsbroek, 2013 | Amsterdam free photos & pictures of the Dutch city | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/104736837@N03/10891535944) |
| `track-28.jpg` | art-desert-dunes-bw | pixellaphoto | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/137643065@N06/23959074279) |
| `track-29.jpg` | Bridge City | Matt Moloney | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://stocksnap.io/photo/bridge-city-9RLAXBXBDP) |
| `track-30.jpg` | Natural Canvas | romainguy | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/), [source image](https://www.flickr.com/photos/24046097@N00/13753901994) |
