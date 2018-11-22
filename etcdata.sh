#!/bin/sh -eu
RANDOM=$$

randdate()
{
	echo $(($(date -d "1 day ago" +%s) + $RANDOM))
}

genname()
{
	echo "SRV-$(echo $RANDOM | tail -c 5)"
}

# Clean previous data
etcdctl rm -r /mutexes

# User lists

etcdctl set /mutexes/user_lists/progress_update/held ""
etcdctl set /mutexes/user_lists/progress_update/owner/hostname `genname`
etcdctl set /mutexes/user_lists/progress_update/owner/since `randdate`
etcdctl set /mutexes/user_lists/progress_update/owner/help "Updates status of the user's progress bar."

etcdctl set /mutexes/user_lists/delete/held ""
etcdctl set /mutexes/user_lists/delete/owner/hostname `genname` 
etcdctl set /mutexes/user_lists/delete/owner/since `randdate`

# User profile

etcdctl set /mutexes/user_profile/inc_visits/held ""
etcdctl set /mutexes/user_profile/inc_visits/owner/hostname `genname`
etcdctl set /mutexes/user_profile/inc_visits/owner/since `randdate`
etcdctl set /mutexes/user_profile/inc_visits/owner/help "Increases visitation counter."

etcdctl set /mutexes/user_profile/dec_visits/held ""
etcdctl set /mutexes/user_profile/dec_visits/owner/hostname `genname`
etcdctl set /mutexes/user_profile/dec_visits/owner/since `randdate`
etcdctl set /mutexes/user_profile/dec_visits/owner/help "Decreases visitation counter."

etcdctl set /mutexes/user_profile/logout/held ""
etcdctl set /mutexes/user_profile/logout/owner/hostname `genname`
etcdctl set /mutexes/user_profile/logout/owner/since `randdate`
etcdctl set /mutexes/user_profile/logout/owner/help "Logs out the user."

etcdctl set /mutexes/user_profile/set_last_login/held ""
etcdctl set /mutexes/user_profile/set_last_login/owner/hostname `genname`
etcdctl set /mutexes/user_profile/set_last_login/owner/since `randdate`

# Billing

etcdctl set /mutexes/billing/pay_cash/held ""
etcdctl set /mutexes/billing/pay_cash/owner/hostname `genname`
etcdctl set /mutexes/billing/pay_cash/owner/since `randdate`
etcdctl set /mutexes/billing/pay_cash/owner/help "Steals money from the user's credit card."

etcdctl set /mutexes/billing/gen_bill/held ""
etcdctl set /mutexes/billing/gen_bill/owner/hostname `genname`
etcdctl set /mutexes/billing/gen_bill/owner/since `randdate`
etcdctl set /mutexes/billing/gen_bill/owner/help "Generate summary of stolen money"

etcdctl set /mutexes/billing/block_payment/held ""
etcdctl set /mutexes/billing/block_payment/owner/hostname `genname`
etcdctl set /mutexes/billing/block_payment/owner/since `randdate`

# Cachepurge

etcdctl set /mutexes/cachepurge/wipe/held ""
etcdctl set /mutexes/cachepurge/wipe/owner/hostname `genname`
etcdctl set /mutexes/cachepurge/wipe/owner/since `randdate`

