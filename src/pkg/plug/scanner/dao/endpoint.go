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
	"fmt"

	"github.com/astaxie/beego/orm"

	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/q"
	"github.com/pkg/errors"
)

func init() {
	orm.RegisterModel(new(models.Endpoint))
}

// AddEndpoint ...
func AddEndpoint(edp *models.Endpoint) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(edp)
}

// GetEndpoint ...
func GetEndpoint(UUID string) (*models.Endpoint, error) {
	e := &models.Endpoint{
		UUID: UUID,
	}

	o := dao.GetOrmer()
	if err := o.Read(e); err != nil {
		return nil, err
	}

	return e, nil
}

// EndpointExists ...
func EndpointExists(UUID string) (bool, error) {
	_, err := GetEndpoint(UUID)

	if err == orm.ErrNoRows {
		return false, nil
	}

	return true, err
}

// UpdateEndpoint ...
func UpdateEndpoint(edp *models.Endpoint, cols ...string) error {
	o := dao.GetOrmer()
	count, err := o.Update(edp, cols...)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no item with UUID %s updated", edp.UUID)
	}

	return nil
}

// DeleteEndpoint ...
func DeleteEndpoint(UUID string) error {
	e := &models.Endpoint{
		UUID: UUID,
	}

	o := dao.GetOrmer()
	count, err := o.Delete(e)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no item with UUID %s deleted", UUID)
	}

	return nil
}

// ListEndpoints ...
func ListEndpoints(q *q.Query) ([]*models.Endpoint, error) {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(models.Endpoint))

	if q != nil {
		if len(q.Keywords) > 0 {
			for k, v := range q.Keywords {
				qt = qt.Filter(fmt.Sprintf("%s__icontains", k), v)
			}
		}

		if q.PageNumber > 0 && q.PageSize > 0 {
			qt = qt.Limit(q.PageSize, (q.PageNumber-1)*q.PageSize)
		}
	}

	l := make([]*models.Endpoint, 0)
	_, err := qt.All(&l)

	return l, err
}
