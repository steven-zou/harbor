// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"
	"github.com/pkg/errors"
)

func init() {
	orm.RegisterModel(new(models.Result))
}

// CreateRecord ...
func CreateRecord(r *models.Result) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(r)
}

// QueryRecord ...
func QueryRecord(digest, endpoint string) (r *models.Result, err error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(new(models.Result))

	r = &models.Result{}
	if err = qs.Filter("endpoint_id", endpoint).
		Filter("digest", digest).
		One(r); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "query record")
	}

	return
}

// GetRecord ...
func GetRecord(id int64) (*models.Result, error) {
	o := dao.GetOrmer()
	r := &models.Result{
		ID: id,
	}

	if err := o.Read(r); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return r, nil
}

// DeleteRecord ...
func DeleteRecord(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.Result{
		ID: id,
	})

	return err
}

// UpdateRecord ...
func UpdateRecord(r *models.Result, cols ...string) error {
	o := dao.GetOrmer()
	count, err := o.Update(r, cols...)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no record with ID %d is updated", r.ID)
	}

	return nil
}

// UpdateRecordStatus ...
func UpdateRecordStatus(trackID int64, status string, statusCode int) error {
	o := dao.GetOrmer()
	r := o.Raw("UPDATE scanner_result SET status = ?, status_code = ? WHERE id = ? AND status_code < ?")
	r = r.SetArgs(status, statusCode, trackID, statusCode)

	_, err := r.Exec()

	return err
}

// GetAllByDigest ...
func GetAllByDigest(digest string) ([]*models.Result, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(new(models.Result))

	all := make([]*models.Result, 0)
	if _, err := qs.Filter("digest", digest).All(&all); err != nil {
		return nil, err
	}

	return all, nil
}
