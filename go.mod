module github.com/qorpress/qorpress

go 1.14

replace github.com/qorpress/qorpress-contrib/oniontree => github.com/qorpress/qorpress/plugins/oniontree v0.0.0-20200310070759-d93c251c2e6e

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/Depado/bfchroma v1.2.0
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/RoaringBitmap/roaring v0.4.21 // indirect
	github.com/Shaked/gomobiledetect v0.0.0-20171211181707-25f014f66568
	github.com/acoshift/paginate v1.1.2
	github.com/alash3al/bbadger v0.0.0-20191001173659-1d440b2b747b // indirect
	github.com/alecthomas/chroma v0.6.7
	github.com/alexedwards/scs v1.4.1
	github.com/aliyun/aliyun-oss-go-sdk v0.0.0-20190307165228-86c17b95fcd5
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496
	github.com/astaxie/beego v1.12.1
	github.com/aws/aws-sdk-go v1.29.18
	github.com/azumads/faker v0.0.0-20150921074035-6cae71ddb107
	github.com/biezhi/gorm-paginator/pagination v0.0.0-20190124091837-7a5c8ed20334
	github.com/blevesearch/bleve v0.8.1
	github.com/blevesearch/go-porterstemmer v1.0.2 // indirect
	github.com/blevesearch/segment v0.0.0-20160915185041-762005e7a34f // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/cep21/xdgbasedir v0.0.0-20170329171747-21470bfc93b9
	github.com/corpix/uarand v0.1.1
	github.com/couchbase/vellum v0.0.0-20190829182332-ef2e028c01fd // indirect
	github.com/dgraph-io/badger v1.6.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.2
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/fatih/color v1.9.0
	github.com/foomo/simplecert v1.6.8
	github.com/foomo/tlsconfig v0.0.0-20180418120404-b67861b076c9
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/goccy/go-yaml v1.4.3
	github.com/gocolly/colly/v2 v2.0.1
	github.com/gohugoio/hugo v0.54.0
	github.com/golang/snappy v0.0.1
	github.com/gomarkdown/markdown v0.0.0-20200127000047-1813ea067497
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v29 v29.0.3
	github.com/gorilla/context v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/gosimple/slug v1.9.0
	github.com/h2non/filetype v1.0.12
	github.com/headzoo/surf v1.0.0
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jinzhu/configor v1.1.1
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/jinzhu/gorm v1.9.12
	github.com/jinzhu/inflection v1.0.0
	github.com/jinzhu/now v1.1.1
	github.com/joho/godotenv v1.3.0
	github.com/jteeuwen/go-bindata v3.0.8-0.20180305030458-6025e8de665b+incompatible
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/karrick/godirwalk v1.15.3
	github.com/koreset/gtf v0.0.0-20180430044607-d9478f26f2ff
	github.com/lib/pq v1.1.1
	github.com/manticoresoftware/go-sdk v0.0.0-20191205035816-0e8dbffac2c9
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-sqlite3 v2.0.1+incompatible
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mrjones/oauth v0.0.0-20190623134757-126b35219450
	github.com/nozzle/throttler v0.0.0-20180817012639-2ea982251481
	github.com/olivere/elastic/v7 v7.0.12
	github.com/onionltd/oniontree-tools v0.0.0-20200217165256-a771af70bf68
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/qiniu/api.v7 v7.2.5+incompatible
	github.com/qiniu/x v7.0.8+incompatible // indirect
	github.com/qorpress/go-wordpress v0.0.0-20200302054333-81c736c8fa04
	github.com/qorpress/grab v2.0.0+incompatible
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/sethgrid/pester v0.0.0-20190127155807-68a33a018ad0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/sniperkit/cacher v0.0.0-20171213170759-61f2a1265daa
	github.com/spf13/pflag v1.0.5
	github.com/steveyen/gtreap v0.0.0-20150807155958-0abe01ef9be2 // indirect
	github.com/tealeg/xlsx v1.0.5
	github.com/theplant/cldr v0.0.0-20190423050709-9f76f7ce4ee8
	github.com/theplant/htmltestingutils v0.0.0-20190423050759-0e06de7b6967
	github.com/theplant/testingutils v0.0.0-20190603093022-26d8b4d95c61
	github.com/unrolled/render v1.0.2
	github.com/x0rzkov/go-vcsurl v1.0.1
	github.com/x0rzkov/httpcache v0.0.0-20200108145149-5a6ae9c4c311
	github.com/yosssi/gohtml v0.0.0-20190915184251-7ff6f235ecaf // indirect
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/bsm/ratelimit.v1 v1.0.0-20160220154919-db14e161995a // indirect
	gopkg.in/h2non/bimg.v1 v1.0.19
	gopkg.in/loremipsum.v1 v1.1.0
	gopkg.in/redis.v3 v3.6.4
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200229041039-0a110f9eb7ab // indirect
	qiniupkg.com/x v7.0.8+incompatible // indirect
)
