/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/2/13
   Description :
-------------------------------------------------
*/

package zcache_broker

type Loger interface {
    Info(v ...interface{})
    Warn(v ...interface{})
    Error(v ...interface{})
}
